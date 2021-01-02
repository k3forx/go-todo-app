CREATE DATABASE IF NOT EXISTS `todo` DEFAULT CHARSET = 'utf8mb4';
USE `todo`;
CREATE TABLE IF NOT EXISTS `tasks` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `user_id` VARCHAR(100) NOT NULL,
  `title` VARCHAR(32) NOT NULL,
  `details` VARCHAR(100) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (
    `id`
  )
) ENGINE = 'InnoDB';

CREATE TABLE IF NOT EXISTS `users` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `user_id` VARCHAR(100) NOT NULL,
  `password` VARCHAR(100) NOT NULL,
  PRIMARY KEY (
    `id`
  )
) ENGINE = 'InnoDB';

INSERT INTO
  tasks (
    user_id,
    title,
    details
  )
VALUES
  ("test_user", "Shopping", "Buy an apple"), ("test_user", "Study", "Read a book");

INSERT INTO
  users (
    user_id,
    password
  )
 VALUES
   ("test_user", "test_password");