-- +migrate Up
create table if not exists device_value (
    -- 值的读取时间。
    value_timestamp timestamp with time zone primary key not null, 
    device_id character varying(32) not null, 
    -- 要存的设备的值,与表004相对应, 有多少个因子就对应多少个值
    -- 这个表商量一下，值是否直接用jsonb存，方便对应关系。
    value_1 decimal,
    value_2 decimal,
    value_3 decimal,
    value_4 decimal,
    value_5 decimal,
    value_7 decimal,
    value_8 decimal,
    value_9 decimal,
    value_10 decimal
);

create unique index idx_value_timestamp on device_value(value_timestamp);

-- +migrate Down
drop index idx_value_timestamp;
drop table device_value;
