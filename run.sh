#!/bin/sh
until mysqladmin ping -h mysql --silent; do
  echo 'waiting for mysqld to be connectable...'
  sleep 2
done

echo "ToDo App is starting...!"
exec go run server.go
