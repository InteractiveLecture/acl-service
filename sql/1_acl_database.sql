

create user aclapp;


create database acl owner aclapp;

\c acl

drop table if exists object_identities cascade;
drop table if exists acl_entries cascade;

create table object_identities (
  id UUID PRIMARY KEY,
  parent_object UUID,
  owner_sid UUID not null,
  constraint fk_acl_obj_parent foreign key(parent_object)references object_identities(id)
);

create table acl_entries (
  object_id UUID not null references object_identities(id),
  sid UUID not null,
  create_permission boolean not null,
  read_permission boolean not null,
  update_permission boolean not null,
  delete_permission boolean not null,
  PRIMARY KEY (object_id,sid)
);




