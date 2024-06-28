use std::collections::HashMap;
use polars::frame::DataFrame;

pub fn df_to_vec(df: DataFrame)-> Vec<HashMap<String,String>>{
    let cols= df.get_column_names();
    let mut res: Vec<HashMap<String,String>>=Vec::new();
    for i in 0..df.shape().0{
        let row=df.get_row(i);
        let mut tmp:HashMap<String, String>=HashMap::new();
        cols.iter().enumerate().for_each(|(col_index,col_name)|{
            let getv=row.0.get(col_index).unwrap();
            tmp.insert(String::from(*col_name), getv.to_string());
        });
        res.push(tmp);
    }
    res
}


use std::fs;
use std::path::Path;
use polars::prelude::IntoLazy;

pub fn get_newest_file(prefix: &str)-> String{
    let dir_path = "./"; // 替换为实际的目录路径

    let file_names = fs::read_dir(dir_path)
        .expect("Failed to read directory")
        .filter_map(|entry| {
            let entry = entry.expect("Failed to read directory entry");
            let path = entry.path();
            let file_name = path.file_name()?.to_string_lossy().into_owned();
            Some((path, file_name))
        })
        .filter(|(_, file_name)| file_name.starts_with(prefix))
        .filter(|(_, file_name)| file_name.ends_with(".csv"))
        .collect::<Vec<_>>();

    let latest_file = file_names
        .into_iter()
        .max_by_key(|(path, _)| {
            path.metadata().unwrap().modified().unwrap()
        });

    if let Some((_, latest_file_name)) = latest_file {
        latest_file_name
    } else {
        "".to_string()
    }
}

//获取两个Vec<String>的交集
pub fn get_union_vec(vecs: Vec<Vec<String>>) -> Vec<String> {
    if let Some(first_vec) = vecs.first() {
        let mut res: Vec<String> = first_vec.clone();

        for vec in vecs.iter().skip(1) {
            res = res.into_iter().filter(|x| vec.contains(x)).collect();
        }

        res
    } else {
        Vec::new()
    }
}


use polars::prelude::*;
//在df中加入ma线（最早的20条记录会混杂其他股票的数据 不准确）
pub fn set_ma(df:&mut DataFrame,i:usize)->Result<(),PolarsError>{
    let   close=df.column("close")?;
    let mut opt=RollingOptionsImpl::default();
    opt.window_size=Duration::new(i as i64);
    opt.min_periods=i;
    let mai=close.reverse().rolling_mean(opt)?.reverse(); //滚动均值
    let col=format!("ma{}",i);
    let mai=Series::new(col.as_str(),mai);

    df.insert_at_idx(0, mai)?;
    Ok(())
}

//在df中加入i日中close最低值
pub fn set_lowest_close(df:&mut DataFrame,i:usize)->Result<(),PolarsError>{
    let   close=df.column("close")?;
    let mut opt=RollingOptionsImpl::default();
    opt.window_size=Duration::new(i as i64);
    opt.min_periods=i;
    let lowest_close=close.reverse().rolling_min(opt)?.reverse(); //滚动均值
    let col=format!("lowest_close_{}",i);
    let lowest_close=Series::new(col.as_str(),lowest_close);

    df.insert_at_idx(0, lowest_close)?;
    Ok(())
}
