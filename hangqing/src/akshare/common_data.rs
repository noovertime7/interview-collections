
use serde::{Serialize,Deserialize};
#[derive(Serialize,Deserialize,Debug,Clone)]
pub struct SimpleStock {
    #[serde(rename = "代码")]
    code: String,
    #[serde(rename = "名称")]
    name: String,
    #[serde(rename = "最新价")]
    price: Option<f64>, //最新价
    #[serde(rename = "总市值")]
    tmv:Option<f64>,//总市值
    #[serde(rename = "流通市值")]
    cmv:Option<f64>,//流通市值
    #[serde(rename = "年初至今涨跌幅")]
    change:Option<f64> // 年初至今的涨跌幅

}
use crate::httphelper::http_get_client;
use anyhow::Result;
use std::collections::HashMap;
use actix_web::Resource;

// 获取所有 股票列表
  async fn get_all_stocks()->Result<Vec<SimpleStock>>{
    let ret_str=http_get_client("stock_zh_a_spot_em",None).await?;
    let ret:Vec<SimpleStock>=serde_json::from_str(ret_str.as_str())?;
    Ok(ret)
}

use csv::Writer;
use polars::prelude::*;
use crate::akshare::helper::{df_to_vec, get_newest_file};

static  TSHEADERS:[&str;6] = ["code","name","price","tmv","cmv","change"];
pub async fn save_all_stocks()->Result<()>{
    let csv_suffix=chrono::Utc::now().format("%Y%m%d").to_string();
    let csv_filename=format!("stocks-{}.csv",csv_suffix);
    let mut  wf=Writer::from_path(csv_filename.clone())?;
    wf.write_record(TSHEADERS)?;// 第一步 先写头 
    
    let data=get_all_stocks().await?;
    data.iter().for_each(|d|{
        wf.write_record([   //写行
            d.code.to_string(),
            d.name.to_string(),
            d.price.unwrap_or(0.0).to_string(),
            d.tmv.unwrap_or(0.0).to_string(),
            d.cmv.unwrap_or(0.0).to_string(),
            d.change.unwrap_or(0.0).to_string()
        ]).unwrap();
    });
    wf.flush()?;
    println!("所有股票列表保存成功,文件是:{}",csv_filename);

    Ok(())
}

use polars::prelude::*;

pub async fn get_stocks_df_from_csv()->Result<DataFrame>{
    let csv_filename = get_newest_file("stocks-");
    let myschema: Schema = Schema::from(
        vec![
            Field::new("code", DataType::Utf8), //强制 把第一列  code 设置为utf8(&str类型)
        ]
    );
    let df = CsvReader::from_path(&csv_filename)?.with_dtypes(Some(&myschema)).finish()?;
    println!("成功读取股票列表文件:{}",csv_filename);
    Ok(df)
}

pub async fn get_all_stocks_from_csv()->Result<Vec<HashMap<String,String>>>{

    let df=get_stocks_df_from_csv().await?;
    let ret = df_to_vec(df);
    Ok(ret)
}

//获得非st股票代码列表
pub async fn get_non_st_code_list()->Result<Vec<String>>{
    let stocks = get_all_stocks_from_csv().await?;
    let codes = stocks.iter().filter(|x| !x.get("name").unwrap().contains("ST")).map(|x| x.get("code").unwrap().to_string()).collect::<Vec<String>>();
    Ok(codes)

}

pub async fn get_all_stocks_code()-> Result<Vec<String>>{

    let df = get_cache(CacheType::AllStocks).await.unwrap();
    let codes = df.column("code")?.utf8()?.
        into_iter().map(|x| x.unwrap().to_string()).collect::<Vec<String>>();
    Ok(codes)
}

use super::stock_cache::*;
use super::control_rate::get_cr_df_from_csv;
// 用于前端显示的一些函数
pub async fn load_cr(ts_code:String)->Result<DataFrame>{
    // 开始读取缓存，如果读到直接返回完事
    let mut df:DataFrame;
    let cache=get_cache(CacheType::ControlRate).await;
    if cache.is_some(){
        println!("从缓存获取cr");
        df=cache.unwrap();

    }else{
        df=get_cr_df_from_csv().await?;
        //插入缓存
        set_cache(CacheType::ControlRate, df.clone()).await;
    }

    if !ts_code.is_empty(){
        Ok(df.lazy().filter(col("code").eq(lit(ts_code))).collect().unwrap())
    }else{ // 显示全部 每个股票显示 前3天
        Ok(df.lazy().groupby([col("code")]).head(Some(3)).limit(1000).collect().unwrap())
    }

}

//加载商誉列表，由于数据太少，直接读文件
pub async fn load_goodwill(ts_code:String)->Result<DataFrame>{
    let csv_file=get_newest_file("goodwill-");
    if csv_file.eq(""){
        return Err(anyhow::Error::msg("未找到商誉文件"));
    }
    let myschema: Schema = Schema::from(
        vec![
            Field::new("code", DataType::Utf8), //强制 把第一列  code 设置为utf8(&str类型)
        ]
    );
    let df=CsvReader::from_path(csv_file)?.
        with_dtypes(Some(&myschema)).finish()?;
    if ts_code.is_empty(){
        return Ok(df);
    }else{
        return  Ok(df.lazy().
            filter(col("code").
                eq(lit(ts_code))).collect()?);
    }

}
// 加载题材概念数据
use super::share_concepts::{get_all_concepts, get_codes_byconcept, get_share_concepts};
pub async fn load_concepts(t:&str, param1:&str)->Result<DataFrame>{
    if t=="1"{  //代表加载所有 概念列表
        let ret= get_all_concepts().await;
        return  ret;
    } else if t=="2"{  //代表加载某个概念的数据 ,获取 搞股票代码和名称
    return  get_codes_byconcept(param1).await;
    }

    return Err(anyhow::Error::msg("不支持的类型"));
}
