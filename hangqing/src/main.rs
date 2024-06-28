
mod akshare;
mod httphelper;

use std::collections::HashMap;
use std::fmt::format;
use akshare::north_founds::*;

// #[tokio::main]
// async fn main()->anyhow::Result<()> {
//     let ret = get_north_founds().await?;
//     println!("{:?}",ret);
//     Ok(())
//
// }

use actix_web::{get, web, App, HttpResponse, HttpServer, Responder, HttpRequest};
use chrono::prelude::*;
use polars::prelude::*;
use akshare::{north_founds::*,control_rate::*};
use qstring::QString;
use akshare::helper::get_union_vec;

#[get("/")]
async fn index() -> impl Responder {
    HttpResponse::Ok().body("Hello, World!")
}

#[get("/stock/nf")]
async fn nf() -> impl Responder {

    let today = NaiveDate::parse_from_str("2023-12-22", "%Y-%m-%d").unwrap();

    let mut df = get_north_founds_df().await.unwrap();

    //筛选
    df = df.lazy().select(&[col("date"), col("value")]).filter(col("date").eq(lit(today))).collect().unwrap();

    //转换成json trait
    let res = akshare::helper::df_to_vec(df);

    web::Json(res)
}

use serde::{Serialize,Deserialize};
#[derive( Serialize,Deserialize,Debug,Clone)]
struct AdminResponse{
    data:Vec<HashMap<String, String>>,
    code:i32
}
#[derive( Serialize,Deserialize,Debug,Clone)]
struct AdminResponseSimple{
    data:Vec<String>,  //主要是这里不同
    code:i32
}

#[derive( Serialize,Deserialize,Debug,Clone)]
struct AdminResponseString{
    data:String,  //主要是这里不同
    code:i32
}

#[get("/admin/commondata/cr")]
async fn ctl(req:HttpRequest) -> impl Responder {

    //获取query参数
    let q=QString::from(req.query_string());
    //获取code参数，默认值是000001
    let ts_code=q.get("ts_code").unwrap();

    let  df = load_cr(ts_code.to_string()).await.unwrap();
    //转换成json trait
    let res = akshare::helper::df_to_vec(df);

    let ret = AdminResponse{
        data:res,
        code:20000
    };

    web::Json(ret)
}

use akshare::common_data::load_goodwill;
#[get("/admin/commondata/goodwill")]  //加载商誉
async fn   admin_commondata_goodwill(req:HttpRequest)->impl Responder{
    let q=QString::from(req.query_string());
    let ts_code=q.get("ts_code").unwrap_or("");
    let   df=load_goodwill(ts_code.to_string()).await.unwrap();
    let vec_data: Vec<HashMap<String, String>>=akshare::helper::df_to_vec(df);
    let  ret=AdminResponse{
        data:vec_data,
        code:20000
    };
    web::Json(ret)
}

use akshare::common_data::load_concepts;
#[get("/admin/commondata/concetps")]  //加载题材概念
async fn   admin_commondata_concepts(req:HttpRequest)->impl Responder{
    let q=QString::from(req.query_string());
    let get_type=q.get("type").unwrap_or("1");
    let param1=q.get("param1").unwrap_or("");
    let   df=load_concepts(get_type,param1).await.unwrap();
    let vec_data: Vec<HashMap<String, String>>=akshare::helper::df_to_vec(df);
    let  ret=AdminResponse{
        data:vec_data,
        code:20000
    };
    web::Json(ret)
}


#[get("/job/save_all_stocks")]
async fn save_all_stocks() -> impl Responder {
    let ret = akshare::common_data::save_all_stocks().await;
    match ret {
        Ok(..) => "保存股票列表文件执行成功".to_string(),
        Err(e) => format!("保存股票列表文件执行成功，但有错:{}", e),
    }
}

use akshare::good_will::save_goodwill;
#[get("/job/save_goodwill")]  //全量抓取商誉
async fn   job_save_goodwill()->impl Responder{
    save_goodwill().await.unwrap();

    "保存股票商誉成功".to_string()
}

use akshare::share_concepts::save_all_share_concepts;
#[get("/job/save_concepts")]  //全量抓取概念
async fn   job_save_concepts()->impl Responder{
    save_all_share_concepts().await.unwrap();

    "保存股票概念成功".to_string()
}

#[get("/stock/list")]
async fn get_all_stocks() -> impl Responder {
    let ret = akshare::common_data::get_all_stocks_from_csv().await;
    match ret {
        Ok(data) => web::Json(data),
        Err(e) => web::Json(Vec::new()),
    }
}

#[get("/stock/list_codes")]
async fn get_all_stock_codes() -> impl Responder {
    let ret = akshare::common_data::get_all_stocks_code().await;
    match ret {
        Ok(data) => web::Json(data),
        Err(e) => web::Json(Vec::new()),
    }
}

use akshare::control_rate::save_all_stock_cr;
#[get("/job/save_cr")]  //全量抓取控盘度 23/
async fn   job_save_cr(req:HttpRequest)->impl Responder{
    let ret=save_all_stock_cr().await;
    if let Err(e)=ret{
        return  format!("保存股票控盘度执行成功,但有错:{}",e);
    }
    "保存股票控盘度成功".to_string()
}

//获取完全控盘的股票代码列表
#[get("/stock/cr_full")]
async fn get_cr_full() -> impl Responder {
    let ret = get_fullctl_codes().await;
    match ret {
        Ok(data) => web::Json(data),
        Err(e) => web::Json(Vec::new()),
    }
}

//保存日k数据---最近一个月
#[get("/job/save_dayk")]
async fn save_dayk(req: HttpRequest) -> impl Responder {
    //获取query参数
    let q=QString::from(req.query_string());
    let period=q.get("period").unwrap_or("month");

    let ret = akshare::day_k::save_stock_dayk(period).await;
    match ret {
        Ok(..) => "保存日K数据执行成功".to_string(),
        Err(e) => format!("保存日K数据执行成功，但有错:{}", e),
    }
}

use akshare::day_k_finder::get_gap_stocks_list;
use crate::akshare::common_data::{get_non_st_code_list, load_cr};
use crate::akshare::day_k_finder::get_gt_ma20_codes;

//筛选完全控盘切近十日有向上跳空且高于ma20的代码列表
#[get("/stock/cr_full_up")]
async fn get_cr_full_up() -> impl Responder {
    let vec_non_st = get_non_st_code_list().await.unwrap();
    let vec_full_ctl = get_fullctl_codes().await.unwrap();
    let vec_gap_n10 = get_gap_stocks_list().await.unwrap();
    let vec_gt_ma20 = get_gt_ma20_codes().await.unwrap();
    let vec_union = get_union_vec(Vec::from([vec_non_st,vec_full_ctl, vec_gap_n10, vec_gt_ma20]));
    println!("同时符合上述条件的股票有{}支", vec_union.len());
    web::Json(vec_union)
}

use akshare::day_k_finder::get_gap_stocks_n10;
#[get("/analysis/stock_dayk1")]  // 有缺口的股票列表
async fn   analysis_stock_dayk1()->impl Responder{
    let codes=get_gap_stocks_n10().await.unwrap();
    let ret=AdminResponseSimple{
        data:codes,
        code:20000

    };
    web::Json(ret)
}

use akshare::day_k_finder::get_keep_rise;
#[get("/analysis/stock_dayk2")]  // 连续上涨的票
async fn   analysis_stock_dayk2()->impl Responder{
    let mut codes=get_keep_rise(5).await.unwrap();
    if codes.len()==0{
        codes.push("没有符合条件的股票".to_string());
    }
    let ret=AdminResponseSimple{
        data:codes,
        code:20000
    };
    web::Json(ret)
}

use akshare::day_k_finder::get_first_drop;
#[get("/analysis/stock_dayk3")]  // 第一次下跌的票
async fn   analysis_stock_dayk3(req: HttpRequest)->impl Responder{
    let q=QString::from(req.query_string());
    let day=q.get("day").unwrap_or("");
    let day_num:u32 = day.parse().unwrap_or(5);
    let codes=get_first_drop(day_num).await.unwrap();
    let ret=AdminResponseSimple{
        data:codes,
        code:20000
    };
    web::Json(ret)
}

use akshare::day_k_finder::get_vol_first_drop;
#[get("/analysis/stock_dayk4")]  // vol柱第一次阴的票
async fn   analysis_stock_dayk4(req: HttpRequest)->impl Responder{
    let q=QString::from(req.query_string());
    let day=q.get("day").unwrap_or("");
    let day_num:u32 = day.parse().unwrap_or(5);
    let codes=get_vol_first_drop(day_num).await.unwrap();
    let ret=AdminResponseSimple{
        data:codes,
        code:20000
    };
    web::Json(ret)
}

use akshare::day_k_finder::get_double_bottom;
#[get("/analysis/stock_dayk5")]  // 双底
async fn   analysis_stock_dayk5(req: HttpRequest)->impl Responder{
    let codes=get_double_bottom().await.unwrap();
    let ret=AdminResponseSimple{
        data:codes,
        code:20000
    };
    web::Json(ret)
}



use actix_cors::*;
use akshare::stock_cache::*;
#[actix_web::main]
async fn main() -> std::io::Result<()> {
    init_cache().await; //初始化缓存
    HttpServer::new(|| {
        App::new()
            .wrap(Cors::default().allow_any_origin().
                allow_any_method().allow_any_header())
            .service(index)
            .service(nf)
            .service(ctl)
            .service(save_all_stocks)
            .service(get_all_stocks)
            .service(get_all_stock_codes)
            .service(job_save_cr)
            .service(get_cr_full)
            .service(save_dayk)
            .service(get_cr_full_up)
            .service(analysis_stock_dayk1)
            .service(job_save_goodwill)
            .service(admin_commondata_goodwill)
            .service(job_save_concepts)
            .service(admin_commondata_concepts)
            .service(analysis_stock_dayk2)
            .service(analysis_stock_dayk3)
            .service(analysis_stock_dayk4)
            .service(analysis_stock_dayk5)
    })
        .bind("127.0.0.1:9090")?
        .run()
        .await
}