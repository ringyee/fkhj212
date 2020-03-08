-- +migrate Up
create table if not exists device_com_conf (
    -- ui界面中相当与那个(采集配置1,相当与设备的id标识)
    device_id character varying(32) primary key not null,
    -- 串口名称等配置参数
    com_interface character varying(20),
	baudrate integer not null,
	bytesize integer not null,
	stopbits integer not null,
	parity character varying(10) not null,
    device_address integer not null,
    protocal character varying(20),
    monitor_type character varying(20)
);

create unique index idx_com_device_id on device_com_conf(device_id);

create table if not exists device_factors (
    -- ui界面中相当与那个(采集配置1,相当与设备的id标识)
    device_id character varying(32) primary key not null,
    -- 因子有关参数
    factor_name character varying(20) not null,
    factor_id integer not null,
    factor_coeffcient integer not null,
    factor_unit character varying(20) not null,
    factor_range_start real not null,
    factor_range_end real not null
);

create unique index idx_factors_device_id on device_factors(device_id);


-- +migrate Down
drop index idx_com_device_id;
drop index idx_factors_device_id;
drop table device_com_conf;
drop table device_factors;
