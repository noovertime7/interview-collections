

//处理十大股东和流通股东
use crate::httphelper::{https_get_client};


use chrono::format;
use serde::{Serialize,Deserialize};
#[derive(Serialize,Deserialize,Debug,Clone)]
pub struct ShareHolderReportDateResponse {
    code:i32,
    success: bool,
    message:String,
    result: Option<ShareHolderReportDateResult>,
}

#[derive(Serialize,Deserialize,Debug,Clone)]
pub struct ShareHolderReportDateResult{
    data:Vec<ShareHolderReportDate>,
    pages:i32,
    count:i32
}
#[derive(Serialize,Deserialize,Debug,Clone)]
pub struct ShareHolderReportDate{
    END_DATE:Option<String>, //报告时间
    IS_REPORTDATE:Option<String>, // 是否报告期    字符串 1 和 0  注意不是数字

}





#[derive(Serialize,Deserialize,Debug,Clone)]
pub struct ShareHolderResponse {
    code:i32,
    success: bool,
    message:String,
    result: Option<ShareHolderDataResult>,
}

#[derive(Serialize,Deserialize,Debug,Clone)]
pub struct ShareHolderDataResult{
    data:Vec<ShareHolderData>,
    pages:i32,
    count:i32
}
#[derive(Serialize,Deserialize,Debug,Clone)]
pub struct ShareHolderData{
    HOLDER_NAME:Option<String>, //股东名称
    HOLD_NUM:Option<i64>, // 持股数
    FREE_HOLDNUM_RATIO:Option<f64>, //占股比例
    HOLD_NUM_CHANGE:Option<String>, //上一年商誉(元)
    #[serde(skip)]
    TS_CODE:Option<String>,  //股票代码
    HOLD_RATIO:Option<f64>// 实控人占股比例
}

use anyhow::Result;
use super::{get_end_date,add_stock_suffix};
//取出最新的报告期
pub async fn get_holderdate(ts_code:&str,is_free:bool)->Result<String> { //取出最新的报告期
    let ts_code=add_stock_suffix(ts_code);
    let mut url=String::new();
    if is_free{
        url=format!("https://datacenter.eastmoney.com/securities/api/data/v1/get?reportName=RPT_F10_EH_FREEHOLDERSDATE&columns=SECUCODE%2CEND_DATE&quoteColumns=&filter=(SECUCODE%3D%22{}%22)(IS_REPORTDATE%3D%221%22)&pageNumber=1&pageSize=5&sortTypes=-1&sortColumns=END_DATE&source=HSF10&client=PC",ts_code);
    }else{
        url=format!("https://datacenter.eastmoney.com/securities/api/data/v1/get?reportName=RPT_F10_EH_HOLDERSDATE&columns=SECUCODE%2CEND_DATE%2CIS_REPORTDATE&quoteColumns=&filter=(SECUCODE%3D%22{}%22)&pageNumber=1&pageSize=50&sortTypes=-1&sortColumns=END_DATE&source=HSF10&client=PC",ts_code);
    }
    let ret_str=https_get_client(url).await?;
    let ret:ShareHolderReportDateResponse=serde_json::from_str(ret_str.as_str())?;
    if ret.success{
          let date_list=ret.result.unwrap().data;
          for date in date_list{
           //十大流通股东 没有 IS_REPORTDATE 字段，取第一个即可
            if date.IS_REPORTDATE.is_none() ||  date.IS_REPORTDATE.unwrap_or("1".to_string())=="1"{
                let datetime_str = date.END_DATE.unwrap();
                let ret_date:Vec<&str>=datetime_str.split_whitespace().collect();
                return Ok(ret_date[0].to_string());
            }
          }
          Err(anyhow::Error::msg(format!("获取十大股东报告期失败1:{}",ret.message)))
    }else{
          Err(anyhow::Error::msg(format!("获取十大股东报告期失败2:{}",ret.message)))
    }

}
//获取十大流通股东
pub async fn get_shareholders(ts_code:&str,is_free:bool) -> Result<Vec<ShareHolderData>> {
    //is_free代表 是否是流通股东
    //默认 每页 500条
   let end_date=get_holderdate(ts_code,is_free).await?;
   let ts_code=add_stock_suffix(ts_code);
   let mut url=String::new();
   if is_free{ //流通股东
       url=format!("https://datacenter.eastmoney.com/securities/api/data/v1/get?reportName=RPT_F10_EH_FREEHOLDERS&columns=SECUCODE%2CSECURITY_CODE%2CEND_DATE%2CHOLDER_RANK%2CHOLDER_NAME%2CHOLDER_TYPE%2CSHARES_TYPE%2CHOLD_NUM%2CFREE_HOLDNUM_RATIO%2CHOLD_NUM_CHANGE%2CCHANGE_RATIO&quoteColumns=&filter=(SECUCODE%3D%22{}%22)(END_DATE%3D%27{}%27)&pageNumber=1&pageSize=&sortTypes=1&sortColumns=HOLDER_RANK&source=HSF10",ts_code,end_date);

   }else{
       url=format!("https://datacenter.eastmoney.com/securities/api/data/v1/get?reportName=RPT_F10_EH_HOLDERS&columns=SECUCODE%2CSECURITY_CODE%2CEND_DATE%2CHOLDER_RANK%2CHOLDER_NAME%2CSHARES_TYPE%2CHOLD_NUM%2CHOLD_NUM_RATIO%2CHOLD_NUM_CHANGE%2CCHANGE_RATIO&quoteColumns=&filter=(SECUCODE%3D%22{}%22)(END_DATE%3D%27{}%27)&pageNumber=1&pageSize=&sortTypes=1&sortColumns=HOLDER_RANK&source=HSF10&client=PC&v=005691059819898747",ts_code,end_date)
   }

    if is_free{
        println!("开始获取十大流通股东");
    }else{
        println!("开始获取十大股东");
    }

    let ret_str=https_get_client(url).await?;
    let ret:ShareHolderResponse=serde_json::from_str(ret_str.as_str())?;
    if ret.success{
          Ok(ret.result.unwrap().data)
    }else{
          Err(anyhow::Error::msg(format!("获取十大股东失败,{}",ret.message)))
    }

}
pub async fn get_holder_master(ts_code:&str)->Result<String>{
    let ts_code=add_stock_suffix(ts_code);
    let url=format!("https://datacenter.eastmoney.com/securities/api/data/v1/get?reportName=RPT_F10_EH_RELATION&columns=SECUCODE%2CSECURITY_CODE%2CHOLDER_NAME%2CHOLD_RATIO&quoteColumns=&filter=(SECUCODE%3D%22{}%22)(RELATED_RELATION%3D%22%E5%AE%9E%E9%99%85%E6%8E%A7%E5%88%B6%E4%BA%BA%22)&pageNumber=1&pageSize=&sortTypes=&sortColumns=&source=HSF10&client=PC",ts_code);
    let ret_str=https_get_client(url).await?;
    let ret:ShareHolderResponse=serde_json::from_str(ret_str.as_str())?;
    if ret.success{
        let ret_data=&ret.result.unwrap().data[0];
          Ok(ret_data.HOLDER_NAME.clone().unwrap())
    }else{
          Err(anyhow::Error::msg(format!("获取{}的实控人失败:{}",ts_code,ret.message)))
    }
}


pub fn save_share_holders(){

}
