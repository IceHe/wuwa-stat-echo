create table wuwa_substat
(
	number integer not null
		constraint wuwa_substats_pk
			primary key,
	name_cn varchar(10) not null,
	bitmap integer,
	name varchar(16)
);

comment on table wuwa_substat is '鸣潮 副词条';

comment on column wuwa_substat.number is '副词条编号';

comment on column wuwa_substat.bitmap is '二进制位图的值：等于 1 << number';

alter table wuwa_substat owner to icehe;

INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (2, '攻击', 4, 'ATK.Rate');
INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (10, '重击', 1024, 'Heavy ATK');
INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (3, '防御', 8, 'DEF.Rate');
INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (5, '攻击固定值', 32, 'ATK.Fixed');
INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (9, '普攻', 512, 'Basic ATK');
INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (11, '共鸣技能', 2048, 'Skill');
INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (0, '暴击', 1, 'Crit.Rate');
INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (4, '生命', 16, 'HP.Rate');
INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (8, '共鸣效率', 256, 'Energy Regen');
INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (12, '共鸣解放', 4096, 'Liberation');
INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (7, '生命固定值', 128, 'HP.Fixed');
INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (1, '暴击伤害', 2, 'Crit.DMG');
INSERT INTO public.wuwa_substat (number, name_cn, bitmap, name) VALUES (6, '防御固定值', 64, 'DEF.Fixed');
