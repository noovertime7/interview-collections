use crate::httphelper::{https_get_client};
 use polars::prelude::{DataFrame, CsvReader, SerReader, IntoLazy};
//最后的数据日期 到时候要适当修改
//https://datacenter-web.eastmoney.com/api/data/v1/get?sortColumns=NOTICE_DATE%2CSECURITY_CODE&sortTypes=-1%2C-1&pageNumber=1&columns=ALL&reportName=RPT_GOODWILL_STOCKDETAILS&filter=(REPORT_DATE%3D%272023-06-30%27)
use serde::{Serialize,Deserialize};
#[derive(Serialize,Deserialize,Debug,Clone)]
struct GoodWillResponse {
    code:i32,
    success: bool,
    message:String,
    result: Option<GoodWillDataResult>,
}
#[derive(Serialize,Deserialize,Debug,Clone)]
struct GoodWillDataResult{
    data:Vec<GoodWillData>,
    pages:i32,
    count:i32
}
#[derive(Serialize,Deserialize,Debug,Clone)]
struct GoodWillData{
    NOTICE_DATE:Option<String>, //公告日期
    SECURITY_CODE:Option<String>, // 股票代码
    SECURITY_NAME_ABBR:Option<String>, //票名称
    GOODWILL_PRE:Option<f64>, //上一年商誉(元)
    GOODWILL:Option<f64>,  //当前商誉(元)
    SUMSHEQUITY_RATIO:Option<f64>  //商誉占净资产比例
}

use anyhow::Result;
use super::{get_end_date};
//商誉相关
async fn get_goodwill(page_number:i32) -> Result<Vec<GoodWillData>> {
    //默认 每页 500条
   let url=format!("https://datacenter-web.eastmoney.com/api/data/v1/get?sortColumns=NOTICE_DATE%2CSECURITY_CODE&sortTypes=-1%2C-1&pageNumber={}&columns=ALL&reportName=RPT_GOODWILL_STOCKDETAILS&filter=(REPORT_DATE%3D%27{}%27)",page_number, get_end_date());

    println!("开始获取商誉,页码 {} ",page_number);
    let ret_str=https_get_client(url).await?;
    let ret:GoodWillResponse=serde_json::from_str(ret_str.as_str())?;
    if ret.success{
          Ok(ret.result.unwrap().data)
    }else{
          Err(anyhow::Error::msg(format!("获取页码 {}的商誉失败,{}",page_number,ret.message)))
    }


}
use chrono_tz::Tz;
use csv::Writer;


static  GW_HEADERS:[&str;6] = ["code","name","goodwill","goodwill_pre","ratio","date"];
//保存商誉文件  格式为goodwill-
pub async fn save_goodwill()->Result<()>{
    let tz: Tz = "Asia/Shanghai".parse().unwrap();
    let csv_suffix=chrono::Utc::now().with_timezone(&tz).format("%Y%m%d").to_string();
    let file_name=format!("goodwill-{}.csv",csv_suffix);

    let mut writer=Writer::from_path(&file_name).unwrap();
    writer.write_record(GW_HEADERS)?;// 写头
    for i in 1..12{ // 12页够了，一般也就6-7页

        let ret=get_goodwill(i).await;
        if let Ok(datalist)=ret{
            for data in datalist{
                let row=[data.SECURITY_CODE.unwrap_or("-".to_string()),
                data.SECURITY_NAME_ABBR.unwrap_or("-".to_string()),
                data.GOODWILL.unwrap_or(0.0).to_string(),
                data.GOODWILL_PRE.unwrap_or(0.0).to_string(),
                data.SUMSHEQUITY_RATIO.unwrap_or(0.0).to_string(),
                chrono::NaiveDate::parse_from_str(data.NOTICE_DATE.unwrap().as_str(),
                    "%Y-%m-%d %H:%M:%S").unwrap().format("%Y-%m-%d").to_string()
               ];
                writer.write_record(row).unwrap();
            }
        }else {
            break; //出错就停
        }

    }
  Ok(())
}
