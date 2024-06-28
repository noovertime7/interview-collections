pub mod getter;
pub use getter::*;


use  serde::{Serialize,Deserialize};
use std::collections::HashMap;
// 基础URL 
static BASE_URL:&str="http://localhost:8080/api/public";
#[derive(Serialize,Deserialize)]
struct AKRequest{   // 请求实体，  
    api_name: String,
    params: HashMap<String,String>,
}