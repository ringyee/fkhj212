-- +migrate Up
create table if not exists up_server_conf (
    -- 相当于ui图中的“上传配置1”
	server_id smallint primary key not null,
	server_address character varying (100) not null,
	server_port smallint  not null,
	device_mn character varying(50) not null,
    protocal varchar(20),
    -- 这界面是设置上报服务地址配置的，上传时间是报文有关的，放这里有什么意思???
    upload_time timestamp with time zone, 
    keeplive integer
);

create unique index idx_server_id on up_server_conf(server_id);

-- +migrate Down
drop index idx_server_id;
drop table up_server_conf;
