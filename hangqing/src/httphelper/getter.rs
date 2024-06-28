use std::collections::HashMap;
use std::str::FromStr;
use anyhow::{Result, Context};
use hyper::{Client,Body,Request,Uri};
use super::BASE_URL;
use hyper_tls::HttpsConnector;

pub async fn https_get_client(url: String) -> Result<String> {
    let https = HttpsConnector::new();
    let client = Client::builder().build::<_, hyper::Body>(https);
    let rsp = client.get(url.parse()?).await?;
    if rsp.status()==200{
        let get_body=hyper::body::to_bytes(rsp.into_body()).await?;

        return  Ok(String::from_utf8(get_body.to_vec())?);

    }else{
        return Err(anyhow::Error::msg("调用失败"));
    }
}

pub async fn http_get_client(api_name: &str, param: Option<HashMap<&str, &str>>) -> Result<String> {

    // 创建一个 Hyper 客户端
    let client = Client::new();

    // 构建请求 URL
    let mut url = format!("{}/{}", BASE_URL, api_name);
    if let Some(param) = param {
        url.push_str("?");
        for (key, value) in param {
            url.push_str(&key);
            url.push_str("=");
            url.push_str(&value);
            url.push_str("&");
        }
        url.pop();
    }
    let uri = Uri::from_str(url.as_str()).context("Failed to parse URI")?;



    // 创建 GET 请求
    let request = Request::builder()
        .uri(uri)
        .method(hyper::Method::GET)
        .body(Body::empty())?;

    // 发送请求并等待响应
    let response = client.request(request).await
        .context("Failed to send request")?;

    // 判定状态码
    if response.status() != 200 {
        return Err(anyhow::Error::msg(format!("Request failed with status code: {}", response.status())));
    }

    // 读取响应体
    let body_bytes = hyper::body::to_bytes(response.into_body()).await
        .context("Failed to read response body")?;
    let body_string = String::from_utf8(body_bytes.to_vec())
        .context("Failed to convert response body to string")?;

    Ok(body_string)
}

