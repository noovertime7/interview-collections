use crate::httphelper::http_get_client;
use serde::{Serialize,Deserialize};
#[derive(Serialize,Deserialize,Debug,Clone)]
pub struct NorthFounds {
    pub date: String,
    pub value: f64,
}


const API_NAME: &str = "stock_hsgt_north_net_flow_in_em";
use anyhow::{Result, Ok};
use std::collections::HashMap;
use url::form_urlencoded;

//默认获取北向总资金
pub async fn get_north_founds() -> Result<Vec<NorthFounds>>{
    let mut params = HashMap::new();
    let encoded =  form_urlencoded::byte_serialize("北上".as_bytes()).collect::<String>();
    params.insert("symbol", encoded.as_str());
    let ret_str: String = http_get_client(API_NAME, Some(params)).await?;
    let ret = serde_json::from_str(ret_str.as_str())?;
    Ok(ret)
}

use chrono::prelude::*;
use polars::prelude::*;

//将 Vec<NorthFounds> 转换成 DataFrame
pub async fn get_north_founds_df() -> Result<DataFrame> {

    let north_founds = get_north_founds().await.unwrap();

    //通过迭代取获取成员vec
    let data_list = north_founds.iter().map(|x| NaiveDate::parse_from_str(x.date.as_str(), "%Y-%m-%d").unwrap()).collect::<Vec<_>>();
    let value_list = north_founds.iter().map(|x| x.value).collect::<Vec<_>>();

    //构建dataFrame
    let df = DataFrame::new(vec![
        Series::new("date", data_list),
        Series::new("value", value_list),
    ])?;
    Ok(df)
}