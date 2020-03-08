-- +migrate Up
create table if not exists base_parameter_set (
    -- 基础参数配置
    factor_name character varying(20) not null,
    factor_id integer not null,
    factor_code integer not null,
    factor_view_name character varying(50) not null
);

create unique index idx_factors_name on base_parameter_set(factor_name);


-- +migrate Down
drop index idx_factors_name;
drop table base_parameter_set;
