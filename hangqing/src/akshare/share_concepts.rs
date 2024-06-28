
// 处理核心题材
use crate::httphelper::{https_get_client};


use chrono::format;
use serde::{Serialize,Deserialize};
#[derive(Serialize,Deserialize,Debug,Clone)]
pub struct ShareConceptResponse {
    code:i32,
    success: bool,
    message:String,
    result: Option<ShareConceptDataResult>,
}

#[derive(Serialize,Deserialize,Debug,Clone)]
pub struct ShareConceptDataResult{
    data:Vec<ShareConceptData>,
    pages:i32,
    count:i32
}

#[derive(Serialize,Deserialize,Debug,Clone)]
pub struct ShareConceptData{
    SECURITY_CODE:Option<String>, //股票代码
    SECURITY_NAME_ABBR:Option<String>, // 股票名称
    SELECTED_BOARD_REASON:Option<String>, // 入选该板块 理由
    NEW_BOARD_CODE:Option<String>,//板块代码
    BOARD_RANK:Option<f64>, //板块排名
    BOARD_NAME:Option<String> //板块名称

}
use super::{add_stock_suffix};
// 根据股票代码获取 题材
pub async fn get_share_concepts(ts_code:&str)->anyhow::Result<Vec<ShareConceptData>> { //取出最新的报告期
    let ts_code=add_stock_suffix(ts_code);
    let   url=format!("https://datacenter.eastmoney.com/securities/api/data/v1/get?reportName=RPT_F10_CORETHEME_BOARDTYPE&columns=SECUCODE%2CSECURITY_CODE%2CSECURITY_NAME_ABBR%2CNEW_BOARD_CODE%2CBOARD_NAME%2CSELECTED_BOARD_REASON%2CIS_PRECISE%2CBOARD_RANK%2CBOARD_YIELD%2CDERIVE_BOARD_CODE&quoteColumns=f3~05~NEW_BOARD_CODE~BOARD_YIELD&filter=(SECUCODE%3D%22{}%22)(IS_PRECISE%3D%221%22)&pageNumber=1&pageSize=&sortTypes=1&sortColumns=BOARD_RANK&source=HSF10&client=PC",ts_code);
    let ret_str=https_get_client(url).await?;
    let ret:ShareConceptResponse=serde_json::from_str(ret_str.as_str())?;
    if ret.success{
        Ok(ret.result.unwrap().data)
    }else{
            Err(anyhow::Error::msg(format!("获取概念失败,{}",ret.message)))
    }
}
use chrono_tz::Tz;
use csv::Writer;
pub fn   get_concepts_file_name()->String{
    let tz: Tz = "Asia/Shanghai".parse().unwrap();
    let csv_suffix=chrono::Utc::now().with_timezone(&tz).format("%Y%m%d").to_string();
    format!("concepts-{}.csv",csv_suffix)
}
use super::common_data::get_all_stocks_code;
static  CONCEPT_HEADERS:[&str;6] = ["code","name","board_code","board_name","board_reason","board_rank"];
pub async fn save_all_share_concepts()->anyhow::Result<()>{
    let   all_stocks=get_all_stocks_code().await?;


    let csv_filename=get_concepts_file_name();
    let mut  wf: Writer<std::fs::File>=Writer::from_path(csv_filename.clone())?;
    wf.write_record(CONCEPT_HEADERS)?;// 第一步 先写头


    for stock in all_stocks.iter(){
        let concepts=get_share_concepts(stock.as_str()).await;
        if concepts.is_err(){
            continue;
        }
        concepts.unwrap().iter().for_each(|d|{
            wf.write_record([   //写行
                d.SECURITY_CODE.clone().unwrap(),
                d.SECURITY_NAME_ABBR.clone().unwrap(),
                d.NEW_BOARD_CODE.clone().unwrap(),
                d.BOARD_NAME.clone().unwrap(),
                d.SELECTED_BOARD_REASON.clone().unwrap(),
                d.BOARD_RANK.clone().unwrap_or(0.0).to_string()
            ]).unwrap();
        });
        wf.flush()?;
    }

    Ok(())
}

use polars::prelude::*;
pub async fn load_concepts_dataframe(file_path:String)->anyhow::Result<DataFrame>{
    let myschema: Schema = Schema::from(
        vec![
            Field::new("code", DataType::Utf8), //强制  code 设置为utf8(&str类型)
        ]
    );
    let   df=CsvReader::from_path(&file_path)?.
    with_dtypes( Some(&myschema)).finish()?;
    Ok(df)
}
use super::stock_cache::{get_cache,CacheType};
pub async fn get_all_concepts()->anyhow::Result<DataFrame>{

   //直接从缓存读
   let dk=get_cache(CacheType::Concepts).await;
   if dk.is_none(){
      return Err(anyhow::anyhow!("题材概念缓存不存在.请预先生成"));
   }
   // select count(board_name),board_name from xxx group by board_name
   let df=dk.unwrap().lazy().groupby([col("board_name")]).agg([
     col("board_name").count().alias("board_name_count"),
   ]).collect()?;
   Ok(df)
}

// 根据题材名称 获取 股票列表  --- 这个很简单
pub async fn get_codes_byconcept(concept:&str)->anyhow::Result<DataFrame>{
    let dk=get_cache(CacheType::Concepts).await;
    if dk.is_none(){
        return Err(anyhow::anyhow!("题材概念缓存不存在.请预先生成"));
    }
    let  concept=concept.replace("\"", "");

    let df=dk.unwrap().lazy().
        filter(col("board_name").str().contains(concept)).
        select([col("code"),col("name"),col("board_reason")]).collect()?;

    Ok(df)
}
