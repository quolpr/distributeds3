-- name: InsertUpload :exec
insert into uploads (id, name, size, status, created_at)
values (@id, @name, @size, @status, @created_at);

-- name: InsertPart :exec
insert into parts (id, server_url, upload_id, number, size, status)
values (@id, @server_url, @upload_id, @number, @size, @status);

-- name: GetUpload :one
select * from uploads where id = @id;

-- name: GetUploadParts :many
select * from parts where upload_id = @id;

-- name: GetOldInProgressUploads :many
select * from uploads where created_at < @created_at and status = 'in_progress';

-- name: GetOldInProgressParts :many
select * from parts where created_at < @created_at and status = 'in_progress';

-- name: DeleteUploadsByIds :exec
delete from uploads where id = ANY(@ids::uuid[]);

-- name: UpdatePartAsDone :exec
update parts set status = 'done' where id = @id;

-- name: UpdateUploadAsDone :exec
update uploads set status = 'done' where id = @id; 
