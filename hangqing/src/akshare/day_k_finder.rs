
//专门放  基于日K进行查找相关的 函数

use std::ops::{Div, Mul};
//加载日K数据，转为dataframe
use anyhow::Result;
use polars::prelude::*;
use crate::akshare::helper::{get_newest_file, set_lowest_close, set_ma};
use crate::akshare::stock_cache::{CacheType, get_cache};
use super::day_k::get_dayk_file_name;
pub async fn load_day_k_df(p: &str) ->Result<DataFrame>{
        let file_name = get_newest_file(format!("dayk-{}-",p).as_str());
        let myschema: Schema = Schema::from(
            vec![
                Field::new("ts_code", DataType::Utf8), //强制   code 设置为utf8(&str类型)
            ]
        );
        let df=CsvReader::from_path(&file_name)?.
         with_dtypes( Some(&myschema)).finish()?;
        println!("成功读取日K文件:{}",file_name);
        Ok(df)
}
// 获取近10日 有向上跳空缺口的 股票
//其实就是 10日内，当日最低>昨日最高的票  low>pre_high
pub async fn get_gap_stocks_list()->Result<Vec<String>>{
   let mut df= load_day_k_df("month").await?;
    //选近10日
   df=df.lazy().groupby([col("ts_code")]).head(Some(10)).collect()?;
    //选出当日最低>昨日最高的票
   df = df.lazy().filter(col("low").gt(col("pre_high"))).collect()?;
   //选出相同code的最近数据
   df = df.lazy().groupby([col("ts_code")]).
          agg([col("*").first()]). //同一个code进行聚合，保留第一条
          select([col("*")]).
          collect()?;


    let codes = df.column("ts_code")?.utf8()?.
        into_iter().map(|x| x.unwrap().to_string()).collect::<Vec<String>>();
    println!("筛选{}支符合条件：近10日出现过向上跳空的票", codes.len());

   return Ok(codes);
}

//获取当日收盘价高于MA20的代码
pub async fn get_gt_ma20_codes()->Result<Vec<String>> {
    let mut df = load_day_k_df("season").await?;

    set_ma(&mut df, 20)?;
    set_ma(&mut df, 10)?;

    //选出相同code的最近数据
    df = df.lazy().groupby([col("ts_code")]).
        agg([col("*").first()]). //同一个code进行聚合，保留第一条
        select([col("*")]).
        collect()?;
    let codes_num = df.shape().0;

    df = df.lazy().filter(col("close").gt(col("ma20"))).collect()?;

    let codes = df.column("ts_code")?.utf8()?.
        into_iter().map(|x| x.unwrap().to_string()).collect::<Vec<String>>();

    println!("筛选{}/{}支符合条件：当日收盘价高于MA20的票", codes.len(), codes_num);
    Ok(codes)
}

// 获取近10日 有缺口，10日线在20日线之上，当日下跌的 股票
pub async fn get_gap_stocks_n10()->Result<Vec<String>>{
    let mut df=get_cache(CacheType::DayKSeason).await.unwrap();
    set_ma(&mut df,10)?;  //计算MA10  ma10
    set_ma(&mut df,20)?;
    let count_condition = |s: DataFrame| -> anyhow::Result<DataFrame, polars::error::PolarsError> {
        let df=s.lazy().
            filter(col("low").gt(col("pre_high"))
                .and(
                    col("low").first().gt(col("ma20").first()) //当前收盘 在20日线之上
                ).and(
                col("ma10").gt(0).and(col("ma20").gt(0)) //有些票 没有ma20
            )
                .and(
                    col("ma10").gt(col("ma20"))//10日线在20日线之上
                ).and(
                col("low").gt(col("ma20")).and(col("low").min().gt(col("ma20").min()))
            ).and(
                col("ts_code").count().gt_eq(10)  //过于新的股不要
            )
                .and(
                    col("close").first().lt(col("open").first())  //只要跌的
                    // col("close").first().lt_eq(col("open").first())  //只要跌的
                )
            ).collect()?;
        Ok(df)
    };
    df=df.lazy().groupby([col("ts_code")]).head(Some(10)).collect()?;
    df=df.lazy().groupby([col("ts_code")]).
        apply(count_condition, Arc::new(Schema::default())).collect()?;

// select ts_code,count(*) from xxx  groupy
    df=df.lazy()
        .groupby([col("ts_code")]).agg([col("ts_code").count().alias("count")])
        .collect()?;
    let codes=df.column("ts_code")?.utf8()?.into_iter().
        map(|x| x.unwrap().to_string()).collect::<Vec<_>>();
    return Ok(codes);
}

// 获取连续涨了day天的票
pub async fn get_keep_rise(day:usize)->Result<Vec<String>>{
    let mut df=get_cache(CacheType::DayKSeason).await.unwrap();
    set_ma(&mut df,5)?;
    set_ma(&mut df,10)?;  //计算MA10  ma10
    set_ma(&mut df,20)?;
    let count_condition = |s: DataFrame| -> anyhow::Result<DataFrame, polars::error::PolarsError> {
        let df=s.lazy()
            .filter((
                // col("low").min().gt_eq(col("ma5").min())  //最低在ma5或之上
                col("low").gt_eq(col("ma5"))  //最低在ma5或之上
            ).and(
                col("ts_code").count().gt_eq(5)
            )
                .and(
                    col("close").gt(col("open"))  //收盘比开盘高 --- 必须是涨的票
                )
                .and(
                    col("ma10").gt(0).and(col("ma20").gt(0)) //有些票 没有ma20
                ).and(
                col("close").first().gt(col("close").last())
            ).and(
                col("ts_code").str().
                    starts_with("60").or(col("ts_code").str().
                    starts_with("300").
                    or(col("ts_code").str().starts_with("68")))
            )
            ).collect()?;



        Ok(df)
    };
    df=df.lazy().groupby([col("ts_code")]).head(Some(day)).collect()?;
    df=df.lazy().groupby([col("ts_code")]).
        apply(count_condition, Arc::new(Schema::default())).collect()?;

    df=df.lazy().groupby([col("ts_code")]).
        agg([col("ts_code").count().alias("count")]).collect().unwrap();

    df=df.lazy().filter(col("count").eq(day as u32)).collect().unwrap();


    let codes=df.column("ts_code")?.utf8()?.into_iter().
        map(|x| x.unwrap().to_string()).collect::<Vec<_>>();
    return Ok(codes);


}

//龙头首阴
pub async fn get_first_drop(day:u32)->Result<Vec<String>> {
    let day_clone = day.clone();
    let mut df=get_cache(CacheType::DayKSeason).await.unwrap();
    let count_condition = move |s: DataFrame| -> anyhow::Result<DataFrame, polars::error::PolarsError> {
        let df=s.lazy()
            .filter(
                col("ts_code").count().gt_eq(day_clone + 1)
                .and(
                    col("close").first().lt(col("open").first()) //当日收盘低于开盘
                ).and(
                    col("close").gt(col("open")) //获取近 n日涨的记录
                ).and(
                    col("ts_code").str().starts_with("60")
                        .or(col("ts_code").str().starts_with("300").
                        or(col("ts_code").str().starts_with("68")))
                )
            ).collect()?;



        Ok(df)
    };
    df=df.lazy().groupby([col("ts_code")]).head(Some(day as usize + 1)).collect()?;
    df=df.lazy().groupby([col("ts_code")]).
        apply(count_condition, Arc::new(Schema::default())).collect()?;

    df=df.lazy().groupby([col("ts_code")]).
        agg([col("ts_code").count().alias("count")]).collect().unwrap();
    df=df.lazy().filter(col("count").eq(day)).collect().unwrap(); //同支股只有一条记录下跌，其他全是阳

    let codes=df.column("ts_code")?.utf8()?.into_iter().
        map(|x| x.unwrap().to_string()).collect::<Vec<_>>();

    println!("现有{}条首阴记录", df.shape().0);

    Ok(codes)
}

pub async fn get_vol_first_drop(day:u32)->Result<Vec<String>> {
    let day_clone = day.clone();
    let mut df=get_cache(CacheType::DayKSeason).await.unwrap();
    let count_condition = move |s: DataFrame| -> anyhow::Result<DataFrame, polars::error::PolarsError> {
        let df=s.lazy()
            .filter(
                col("ts_code").count().gt_eq(day_clone + 1)
                    .and(
                        col("close").first().lt(col("pre_close").first()) //当日收盘低于昨日收盘
                    ).and(
                    col("close").gt(col("pre_close")) //获取近 n日涨的记录
                ).and(
                    col("ts_code").str().starts_with("60")
                        .or(col("ts_code").str().starts_with("300").
                            or(col("ts_code").str().starts_with("68")))
                )
            ).collect()?;



        Ok(df)
    };
    df=df.lazy().groupby([col("ts_code")]).head(Some(day as usize + 1)).collect()?;
    df=df.lazy().groupby([col("ts_code")]).
        apply(count_condition, Arc::new(Schema::default())).collect()?;

    df=df.lazy().groupby([col("ts_code")]).
        agg([col("ts_code").count().alias("count")]).collect().unwrap();
    df=df.lazy().filter(col("count").eq(day)).collect().unwrap(); //同支股只有一条记录下跌，其他全是阳

    let codes=df.column("ts_code")?.utf8()?.into_iter().
        map(|x| x.unwrap().to_string()).collect::<Vec<_>>();

    println!("现有vol{}条首阴记录", df.shape().0);

    Ok(codes)
}

//获取出现了双底的股票
pub async fn get_double_bottom() ->Result<Vec<String>>{
    let mut df = load_day_k_df("season").await?;
    //添加最低收盘价的列
    set_lowest_close(&mut df, 20)?;
    set_lowest_close(&mut df, 10)?;

    let count_condition = move |s: DataFrame| -> anyhow::Result<DataFrame, polars::error::PolarsError> {
        let df=s.lazy()
            .filter(
                col("lowest_close_20").first().neq(col("lowest_close_10").first())
                    .and(//20日最低收盘价小于10日最低收盘价的102%且大于10日最低收盘价的98%
                         col("lowest_close_20").first().div(col("lowest_close_10").first()).lt(1.02)
                    )
                    .and(
                         col("lowest_close_20").first().div(col("lowest_close_10").first()).gt(0.98)
                    )
                    .and(
                        col("ts_code").str().starts_with("60")
                            .or(col("ts_code").str().starts_with("300").
                                or(col("ts_code").str().starts_with("68")))
                    )
            ).collect()?;



        Ok(df)
    };

    df=df.lazy().groupby([col("ts_code")]).
        apply(count_condition, Arc::new(Schema::default())).collect()?;

    df=df.lazy()
        .groupby([col("ts_code")]).agg([col("ts_code").count().alias("count")])
        .collect()?;
    let codes=df.column("ts_code")?.utf8()?.into_iter().
        map(|x| x.unwrap().to_string()).collect::<Vec<_>>();

    Ok(codes)
}