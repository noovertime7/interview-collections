
pub mod north_founds;
pub mod helper;
pub mod control_rate;
pub(crate) mod common_data;
pub(crate) mod day_k;
pub mod day_k_finder;

pub mod stock_cache;
pub mod share_concepts;
pub mod share_holder;
pub mod good_will;

use north_founds::*;

pub fn add_stock_suffix(code: &str) -> String {
    let prefix = &code[..2];
    match prefix {
        "60" | "68" => format!("{}.SH", code),
        "00" | "300" => format!("{}.SZ", code),
        _ => code.to_owned(),
    }
}


//  有些接口 要填，以后大家都调用这个
pub fn get_end_date() -> String {
    "2023-06-30".to_string()
}