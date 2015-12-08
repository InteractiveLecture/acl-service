with recursive p(id,level,path,object_class,parent_object) as (
  select id,0, '/'||id ,object_class , parent_object from object_identities where parent_object is null
  union
  select o.id,p.level+1, p.path || '/' || o.id, o.object_class,o.parent_object 
  from object_identities o inner join p on p.id= o.parent_object
) select id,object_class,path from p;
