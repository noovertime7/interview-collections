
use moka::future::Cache;
// github地址:https://github.com/moka-rs/moka
//抄抄的。没啥好研究的


use once_cell::sync::OnceCell;
use polars::prelude::DataFrame;
use crate::akshare::common_data::{get_all_stocks_from_csv, get_stocks_df_from_csv};
use crate::akshare::control_rate::get_cr_df_from_csv;
use crate::akshare::day_k_finder::load_day_k_df;
use crate::akshare::helper::get_newest_file;
use super::share_concepts::load_concepts_dataframe;

static STOCK_CACHE:OnceCell<Cache<CacheType,DataFrame>>=OnceCell::new();

//外部的main函数  要调用此函数初始化
pub async fn init_cache(){
    //64 * 1024 * 1024   代表有64M的内存最大
    STOCK_CACHE.set(Cache::new(64 * 1024 * 1024)).unwrap();

    //预先插入 一些肯定要用到的缓存
    //1、加载控盘度
    let cr_df_ret=get_cr_df_from_csv().await;
    if let Ok(cr_df)=cr_df_ret{
        set_cache(CacheType::ControlRate, cr_df.clone()).await;
    } else {
        println!("加载控盘度失败");
    }

    //2、加载所有股票列表
    let stocks_df_ret = get_stocks_df_from_csv().await;
    if let Ok(stocks_df)=stocks_df_ret{
        set_cache(CacheType::AllStocks, stocks_df.clone()).await;
    } else {
        println!("加载所有股票列表失败");
    }

    //3、加載日K緩存(3个月的)
    let dayk_df_ret = load_day_k_df("season").await;
    if let Ok(dayk_df)=dayk_df_ret{
        set_cache(CacheType::DayKSeason, dayk_df.clone()).await;
    } else {
        println!("加载日K缓存失败");
    }

    // 4、加载题材概念数据
    let concepts_file=get_newest_file("concepts-");
    if !concepts_file.eq(""){
        let  df_ret=load_concepts_dataframe(concepts_file).await;
        if  df_ret.is_ok(){
            set_cache(CacheType::Concepts, df_ret.unwrap()).await;
            println!("概念数据加载成功");
        }else{
            println!("概念数据加载失败");
        }
    }
}
#[derive(Debug, PartialEq, Eq, Hash)]
pub enum CacheType {
    ControlRate, //控盘度
    AllStocks,// 所有股票
    DayKSeason, //日K(3个月的)
    Concepts,//概念题材
}
pub async fn get_cache(key:CacheType)->Option<DataFrame>{
    STOCK_CACHE.get().unwrap().get(&key).await
}
pub async fn set_cache(key:CacheType,value:DataFrame){
    STOCK_CACHE.get().unwrap().insert(key,value).await;
}
