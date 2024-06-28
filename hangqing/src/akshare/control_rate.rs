use std::fmt::format;
use serde::{Serialize, Deserialize};
#[derive(Serialize,Deserialize,Debug,Clone)]
struct CtlResponse {
    code:i32,
    version:String,
    success: bool,
    message:String,
    result: CtlDataResult,
}
#[derive(Serialize,Deserialize,Debug,Clone)]
struct CtlDataResult{
    data:Vec<CtlData>,
}
#[derive(Serialize,Deserialize,Debug,Clone)]
struct CtlData{
    PRIME_COST_20DAYS:Option<f64>, //20日主力成本
    PRIME_COST_60DAYS:Option<f64>, // 60日主力成本
    ORG_PARTICIPATE:Option<f64>, //控盘度
    PARTICIPATE_TYPE_CN:Option<String>, //控盘度描述
    TRADE_DATE:String  //日期
}

use anyhow::{Result};
use crate::httphelper::https_get_client;

async fn get_control_rate(tsCode: &str)-> Result<Vec<CtlData>>{
    let url =  format!("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_DMSK_TS_STOCKEVALUATE&filter=(SECURITY_CODE%3D%22{}%22)&columns=ALL&sortColumns=TRADE_DATE&sortTypes=-1", tsCode);
    println!("开始获取code={}的控盘度",tsCode);
    let res = https_get_client(url).await?;
    let rsp: CtlResponse = serde_json::from_str(res.as_str())?;
    if rsp.success {
        Ok(rsp.result.data)
    }else{
        Err(anyhow::Error::msg(rsp.message))
    }
}

use chrono::prelude::*;
use polars::prelude::*;

//将 Vec<CtlData> 转换成 DataFrame
pub async fn get_ctl_data_df(tscode: &str) -> Result<DataFrame> {

    let data=get_control_rate(tscode).await?;

    //通过迭代取获取成员vec

    let trade_list=data.iter().map(|x| NaiveDate::parse_from_str(x.TRADE_DATE.as_str(), "%Y-%m-%d %H:%M:%S").unwrap()).collect::<Vec<_>>();
    let c20_list= data.iter().map(|x| x.PRIME_COST_20DAYS.unwrap_or(0.0)).collect::<Vec<_>>();
    let c60_list: Vec<f64>= data.iter().map(|x| x.PRIME_COST_60DAYS.unwrap_or(0.0)).collect::<Vec<_>>();
    let rate_list: Vec<f64>= data.iter().map(|x| x.ORG_PARTICIPATE.unwrap_or(0.0)).collect::<Vec<_>>();
    let rate_cn_list:Vec<String>=data.iter().map(|x| x.PARTICIPATE_TYPE_CN.clone().
        unwrap_or("-".to_string())).collect::<Vec<_>>();

    //构建dataFrame
    let mut df = DataFrame::new(vec![
        Series::new("trade_date", trade_list),
        Series::new("c20", c20_list),
        Series::new("c60", c60_list),
        Series::new("rate", rate_list),
        Series::new("rate_cn", rate_cn_list),
    ])?;

    //根据trade_data倒序排序
    let sort_option = SortOptions{
        descending: true,
        nulls_last: false
    };
    df=df.lazy().sort("trade_date", sort_option).collect()?;

    Ok(df)
}

use crate::akshare::common_data;
use std::sync::Mutex;
use futures::stream::{FuturesUnordered};
use futures::future::join_all;
use tokio::sync::Semaphore;

static  CR_HEADERS:[&str;6] = ["code","c20","c60","rate","cn","date"];
//保存所有股票的控盘度
pub async fn save_all_stock_cr()->Result<()>{
    let codes=common_data::get_all_stocks_code().await?;

    let csv_suffix=chrono::Utc::now().format("%Y%m%d").to_string();
    let csv_filename=format!("cr-{}.csv",csv_suffix);
    //删除文件
    std::fs::remove_file(csv_filename.clone()).unwrap_or(());

    let mut  wf=Arc::new(Mutex::new(csv::Writer::from_path(csv_filename.clone())?));
    wf.lock().unwrap().write_record(CR_HEADERS)?;// 第一步 先写头

    let futures = FuturesUnordered::new();
    let max_concurrent_tasks = 20; // 你想要限制的最大任务数
    let semaphore = Arc::new(Semaphore::new(max_concurrent_tasks));

    for code in codes{
        if code.starts_with("68") || code.starts_with("30") || code.starts_with("60") || code.starts_with("00"){
            let sem_clone = Arc::clone(&semaphore);
            let wf_clone = wf.clone();
            let task = tokio::spawn(async move {
                let _permit = sem_clone.acquire().await.unwrap(); //准入
                let rates_ret = get_control_rate(&code).await;
                if let Ok(rates) = rates_ret{
                    for rate in rates{
                        wf_clone.lock().unwrap().write_record([   //写行
                            code.to_string(),
                            rate.PRIME_COST_20DAYS.unwrap_or(0.0).to_string(),
                            rate.PRIME_COST_60DAYS.unwrap_or(0.0).to_string(),
                            rate.ORG_PARTICIPATE.unwrap_or(0.0).to_string(),
                            rate.PARTICIPATE_TYPE_CN.unwrap_or("-".to_string()),
                            NaiveDate::parse_from_str(rate.TRADE_DATE.as_str(), "%Y-%m-%d %H:%M:%S").unwrap().format("%Y-%m-%d").to_string()
                        ]).expect("write record error");
                    }
                }
            });
            futures.push(task);
        }
    }
    join_all(futures).await;
    println!("所有股票控盘度保存成功,文件是:{}",csv_filename);
    wf.lock().unwrap().flush()?;
    Ok(())
}

use crate::akshare::helper::get_newest_file;

//从csv文件读取cr dataframe
pub async fn get_cr_df_from_csv()->Result<DataFrame>{
    let csv_filename = get_newest_file("cr-");
    if csv_filename.eq(""){
        return Err(anyhow::Error::msg(format!("控盘度文件不存在.请调用/job/save_cr生成")));
    }

    //将第一列的code设置为&str类型
    let myschema = Schema::from(vec![
        Field::new("code", DataType::Utf8),
    ]);
    let mut df = CsvReader::from_path(csv_filename.clone())?
        .infer_schema(None)
        .has_header(true)
        .with_dtypes(Some(&myschema))
        .finish()?;

    println!("成功读取控盘度文件:{}",csv_filename);

    //根据code和date排序
    let sort_option = SortOptions{
        descending: true,
        nulls_last: false
    };
    df=df.lazy().select([col("*")]).
        sort("code", sort_option).
        sort("date", sort_option).collect()?;

    Ok(df)
}

//获取完全控盘的股票
pub async fn get_fullctl_codes()->Result<Vec<String>>{
    let df = get_cr_df_from_csv().await?;
    let mut df_aggr = df.lazy().groupby([col("code")]).
        agg([col("*").first()]). //同一个code进行聚合，保留第一条
        select([col("*")]).
        collect()?;
    //获取df的行数
    let row_count = df_aggr.shape().0;
    df_aggr = df_aggr.lazy().filter(col("cn").eq(lit("完全控盘"))).collect()?;

    let codes = df_aggr.column("code")?.utf8()?.
        into_iter().map(|x| x.unwrap().to_string()).collect::<Vec<String>>();
    println!("筛选{}/{}支符合条件：完全控盘的记录", codes.len(), row_count);
    Ok(codes)
}