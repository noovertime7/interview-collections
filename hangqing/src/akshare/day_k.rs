
///日K 数据获取  基于- akshare . tushare 每天只有2万次调用
//文档在这 https://akshare.akfamily.xyz/data/stock/stock.html#id20
//http://localhost:8080/api/public/stock_zh_a_hist?symbol=002375&start_date=20230301&adjust=qfq


use serde::{Serialize,Deserialize};
#[derive(Serialize,Deserialize,Debug,Clone)]
pub struct DayK{
    #[serde(rename = "日期")]
    trate_date: String,
    #[serde(skip_deserializing)]  //原始 数据中没有，后期人工赋值的
    code: String,
    #[serde(rename = "开盘")]
    open: Option<f64>,
    #[serde(rename = "收盘")]
    close: Option<f64>, 
    #[serde(skip_deserializing)]  //原始 数据中没有，后期人工计算的
    pre_close: Option<f64>, 
    #[serde(rename = "最高")]
    high:Option<f64>,
    #[serde(skip_deserializing)] //昨日最高
    pre_high:Option<f64>,
    #[serde(rename = "最低")]
    low:Option<f64>,
    #[serde(rename = "skip_deserializing")]  //昨日最低
    pre_low:Option<f64>,
    #[serde(rename = "成交量")]
    volume:Option<f64>,
    #[serde(rename = "成交额")]
    tvolume:Option<f64>,
    #[serde(rename = "振幅")]
    amplitude:Option<f64>,
    #[serde(rename = "涨跌幅")]
    pct_change:Option<f64>,
    #[serde(rename = "涨跌额")]
    change:Option<f64>,
    #[serde(rename = "换手率")]
    trate:Option<f64>,
}
static  K_HEADERS:[&str;15] = ["trade_date","ts_code","open","close","pre_close","high","pre_high","low","pre_low",
"volume","tvolume","amplitude","pct_change","change","trate"];

use std::collections::HashMap;
use chrono::NaiveDate;

use anyhow::Result;
use crate::httphelper::http_get_client;
// 根据股票代码 获取该股票的日K
pub async fn get_stock_dayk(code:&str,start_date:NaiveDate,end_date:NaiveDate)->Result<Vec<DayK>>{
    let mut  params:HashMap<&str,&str>=HashMap::new();
     let start_date_str=start_date.format("%Y%m%d").to_string();
     let end_date_str=end_date.format("%Y%m%d").to_string();
    params.insert("symbol", code);
    params.insert("start_date", start_date_str.as_str());
    params.insert("end_date", end_date_str.as_str());
    params.insert("adjust", "qfq");
    let ret_str=http_get_client("stock_zh_a_hist",Some(params)).await?;
    let mut ret:Vec<DayK>=serde_json::from_str(ret_str.as_str())?;
    ret.iter_mut().for_each(|x|{
        x.code=code.to_string();
        let parts: Vec<&str> = x.trate_date.split('T').collect();
        x.trate_date=parts[0].to_string();
    });
    ret.reverse(); // 接口是按 日期正排序的，我们倒过来

    // 处理pre_close、pre_low 原来是没值的
    if  ret.len()>0{
        for i in (0..ret.len() - 1).rev() {
            ret[i].pre_close = ret[i + 1].close;
            ret[i].pre_low = ret[i + 1].low;
            ret[i].pre_high = ret[i + 1].high;
        }
    }
    
    Ok(ret)
}

use super::common_data::get_all_stocks_code;
use std::sync::{Mutex,Arc};
use csv::Writer;
pub(crate) fn get_dayk_file_name(p:&str) ->String{
 
    let csv_suffix=chrono::Utc::now().format("%Y%m%d").to_string();
    format!("dayk-{}-{}.csv",p,csv_suffix)
}
use tokio::sync::Semaphore;
use chrono::Utc;
use chrono::Duration;
//生成所有股票一个月内的日K数据
pub async fn save_stock_dayk(period: &str) ->Result<()>{
   let codes=get_all_stocks_code().await.unwrap();
   let file_name=get_dayk_file_name(period);

   let current_date = Utc::now().date_naive();
    let day_count = match period {
        "month" => 30,
        "season" => 90,
        "year" => 365,
        _ => 30,
    };
   // 获取最近一个周期的日期
   let previous_date = Utc::now().naive_utc() - Duration::days(day_count);
   let wf: Arc<Mutex<Writer<std::fs::File>>> = Arc::new(Mutex::new(Writer::from_path(&file_name)?));
   wf.lock().unwrap().write_record(K_HEADERS)?;// 写头

   let max_concurrent_tasks = 20; // 你想要限制的最大任务数
   let semaphore = Arc::new(Semaphore::new(max_concurrent_tasks));
   
  let joinhandlers= codes.into_iter().map(|x|{
         let wf_clone=wf.clone();
         let sem_clone = Arc::clone(&semaphore);
        return tokio::spawn(async move {
           
              let _permit = sem_clone.acquire().await.unwrap();
                let ret=get_stock_dayk( x.as_str(),previous_date.date()
                ,current_date,
                ).await;
                match ret {
                  Ok(dayks)=>{
                        for dayk in dayks{
                            wf_clone.lock().unwrap().write_record(
                               [dayk.trate_date,
                               dayk.code,
                               dayk.open.unwrap_or(0.0).to_string(), 
                               dayk.close.unwrap_or(0.0).to_string(),  
                               dayk.pre_close.unwrap_or(0.0).to_string(),  
                               dayk.high.unwrap_or(0.0).to_string(),
                               dayk.pre_high.unwrap_or(0.0).to_string(),
                               dayk.low.unwrap_or(0.0).to_string(),
                               dayk.pre_low.unwrap_or(0.0).to_string(),  
                               dayk.volume.unwrap_or(0.0).to_string(),  
                               dayk.tvolume.unwrap_or(0.0).to_string(),  
                               dayk.amplitude.unwrap_or(0.0).to_string(),  
                               dayk.pct_change.unwrap_or(0.0).to_string(),
                               dayk.change.unwrap_or(0.0).to_string(),    
                               dayk.trate.unwrap_or(0.0).to_string(),  
                               ]
                            ).unwrap();   
                        }
                        return Ok(());
                    }
                    Err(e)=>{
                        return Err(anyhow::anyhow!(e));
                    }
                }  

        });

   }).collect::<Vec<_>>();
   let join_ret=tokio::join!(async{joinhandlers});
   for ret in join_ret.0{
        ret.await.unwrap();
     }
  Ok(())
}