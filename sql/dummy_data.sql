CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
truncate table object_identities cascade;
truncate table acl_entries cascade;


insert into object_identities(id,object_class,parent_object,owner_sid) values (
  uuid_generate_v3(uuid_ns_url(),'topic_1'),
  'TOPIC',
  null,
  uuid_generate_v3(uuid_ns_url(),'admin'));

insert into object_identities(id,object_class,parent_object,owner_sid) values (
  uuid_generate_v3(uuid_ns_url(),'module_1'),
  'MODULE',
  uuid_generate_v3(uuid_ns_url(),'topic_1'),
  uuid_generate_v3(uuid_ns_url(),'officer_1'));

insert into object_identities(id,object_class,parent_object,owner_sid) values (
  uuid_generate_v3(uuid_ns_url(),'exercise_1'),
  'EXERCISE',
  uuid_generate_v3(uuid_ns_url(),'module_1'),
  uuid_generate_v3(uuid_ns_url(),'assistant_1'));

insert into acl_entries(object_id,sid,create_permission,read_permission,update_permission,delete_permission)
values (uuid_generate_v3(uuid_ns_url(),'topic_1'),uuid_generate_v3(uuid_ns_url(),'admin'),true,true,true,true);

insert into acl_entries(object_id,sid,create_permission,read_permission,update_permission,delete_permission)
values (uuid_generate_v3(uuid_ns_url(),'topic_1'),uuid_generate_v3(uuid_ns_url(),'officer_1'),true,true,true,false);


insert into acl_entries(object_id,sid,create_permission,read_permission,update_permission,delete_permission)
values (uuid_generate_v3(uuid_ns_url(),'module_1'),uuid_generate_v3(uuid_ns_url(),'officer_1'),true,true,true,true);

insert into acl_entries(object_id,sid,create_permission,read_permission,update_permission,delete_permission)
values (uuid_generate_v3(uuid_ns_url(),'exercise_1'),uuid_generate_v3(uuid_ns_url(),'assistant_1'),true,true,true,true);
