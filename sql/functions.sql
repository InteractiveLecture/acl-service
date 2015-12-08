drop function get_parent_tree(UUID);
CREATE OR REPLACE FUNCTION get_parent_tree(in_id UUID) 
RETURNS table(id UUID, level int, object_class varchar(500), path text ,parent_object UUID) AS $$
with recursive p(id,level,path,object_class,parent_object) as (
  select id,0, '/'||id ,object_class , parent_object from object_identities where id = in_id
  union
  select o.id,p.level+1, p.path || '/' || o.id, o.object_class,o.parent_object 
  from object_identities o inner join p on p.parent_object = o.id
) select id,level,object_class,path,parent_object from p;
$$ LANGUAGE sql;

drop function get_permissions(UUID,UUID);
drop type permission;
create type permission as (
  object_id         UUID,
  sid               UUID,
  read_permission   boolean,
  create_permission boolean,
  update_permission boolean,
  delete_permission boolean
);

CREATE OR REPLACE FUNCTION get_permissions(in_object_id UUID, in_sid UUID) 
RETURNS json AS $$
DECLARE
result permission;
tmp record;
BEGIN
  result.object_id = in_object_id;
  result.sid = in_sid;
  result.read_permission = false;
  result.create_permission = false;
  result.update_permission = false;
  result.delete_permission = false;
  FOR tmp IN select * from get_parent_tree(in_object_id) t inner join acl_entries ac on t.id = ac.object_id where ac.sid = in_sid LOOP
    RAISE NOTICE 'got record with id % and values % % % %', tmp.object_id, tmp.read_permission, tmp.create_permission,tmp.update_permission,tmp.delete_permission;
    if tmp.read_permission THEN
      result.read_permission = true;
    END IF;
    IF tmp.create_permission THEN
      result.create_permission = true;
    END IF;
    IF tmp.update_permission THEN
      result.update_permission = true;
    END IF;
    IF tmp.delete_permission THEN
      result.delete_permission = true;
    END IF;
END LOOP;
return to_json(result);
END;
$$ LANGUAGE plpgsql;

