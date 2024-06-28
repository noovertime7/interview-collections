-- 创建 user 表
CREATE DATABASE IF NOT EXISTS test;
use test;CREATE TABLE IF NOT EXISTS user (
                      id INT PRIMARY KEY,
                      name VARCHAR(255)
);

-- 插入示例数据
INSERT INTO users (id, name) VALUES (1, 'John');
INSERT INTO users (id, name) VALUES (2, 'Jane');
INSERT INTO users (id, name) VALUES (3, 'Alice');
