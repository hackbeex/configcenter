/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

# Create Database
# ------------------------------------------------------------
CREATE DATABASE IF NOT EXISTS cc_config DEFAULT CHARACTER SET = utf8mb4;

Use cc_config;

# Dump of table app
# ------------------------------------------------------------

DROP TABLE IF EXISTS app;

CREATE TABLE app (
  id CHAR(36) NOT NULL COMMENT '',
  name VARCHAR(64) NOT NULL COMMENT 'uniqueness name in cluster',
  comment VARCHAR(255) NOT NULL DEFAULT '' COMMENT '',
  is_delete TINYINT(1) NOT NULL DEFAULT 0 COMMENT '',
  create_by CHAR(32) NOT NULL COMMENT '',
  create_time INT NOT NULL COMMENT '',
  update_by CHAR(32) DEFAULT '' COMMENT '',
  update_time INT NULL COMMENT '',
  PRIMARY KEY (id),
  KEY idx_name (name),
  KEY idx_update_time (update_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='';



# Dump of table namespace
# ------------------------------------------------------------

DROP TABLE IF EXISTS namespace;

CREATE TABLE namespace (
  id CHAR(36) NOT NULL COMMENT '',
  name VARCHAR(64) NOT NULL COMMENT 'uniqueness name in app',
  app_id CHAR(32) NOT NULL  COMMENT '',
  cluster_id CHAR(32) NOT NULL COMMENT '',
  is_public TINYINT(1) NOT NULL DEFAULT 0 COMMENT '',
  comment VARCHAR(64) NOT NULL DEFAULT '' COMMENT '',
  is_delete TINYINT(1) NOT NULL DEFAULT 0 COMMENT '',
  create_by CHAR(32) NOT NULL COMMENT '',
  create_time INT NOT NULL COMMENT '',
  update_by CHAR(32) DEFAULT '' COMMENT '',
  update_time INT NULL COMMENT '',
  PRIMARY KEY (id),
  KEY idx_name (name),
  KEY idx_app_id (app_id),
  KEY idx_cluster_id (cluster_id),
  KEY idx_update_time (update_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='';



# Dump of table record
# ------------------------------------------------------------

DROP TABLE IF EXISTS record;

CREATE TABLE record (
  id INT(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '',
  table_name VARCHAR(32) NOT NULL COMMENT 'db table name',
  table_id VARCHAR(36) NOT NULL COMMENT '',
  op_type VARCHAR(32) NOT NULL COMMENT 'operation type',
  comment VARCHAR(255) DEFAULT '' COMMENT '',
  is_delete TINYINT(1) NOT NULL DEFAULT 0 COMMENT '',
  create_by CHAR(32) NOT NULL COMMENT '',
  create_time INT NOT NULL COMMENT '',
  update_by CHAR(32) DEFAULT '' COMMENT '',
  update_time INT NULL COMMENT '',
  PRIMARY KEY (id),
  KEY idx_update_time (update_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='db operation record';



# Dump of table cluster
# ------------------------------------------------------------

DROP TABLE IF EXISTS cluster;

CREATE TABLE cluster (
  id CHAR(36) NOT NULL COMMENT '',
  name VARCHAR(64) NOT NULL COMMENT 'uniqueness name in dev',
  app_id CHAR(32) NOT NULL COMMENT '',
  comment VARCHAR(255) NOT NULL DEFAULT '' COMMENT '',
  is_delete TINYINT(1) NOT NULL DEFAULT 0 COMMENT '',
  create_by CHAR(32) NOT NULL COMMENT '',
  create_time INT NOT NULL COMMENT '',
  update_by CHAR(32) DEFAULT '' COMMENT '',
  update_time INT NULL COMMENT '',
  PRIMARY KEY (id),
  KEY idx_name (name),
  KEY idx_app_id (app_id),
  KEY idx_update_time (update_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='';



# Dump of table commit
# ------------------------------------------------------------

DROP TABLE IF EXISTS commit;

CREATE TABLE commit (
  id CHAR(36) NOT NULL COMMENT '',
  namespace_id CHAR(32) NOT NULL COMMENT '',
  change_sets LONGTEXT NOT NULL COMMENT '',
  is_delete TINYINT(1) NOT NULL DEFAULT 0 COMMENT '',
  create_by CHAR(32) NOT NULL COMMENT '',
  create_time INT NOT NULL COMMENT '',
  update_by CHAR(32) DEFAULT '' COMMENT '',
  update_time INT NULL COMMENT '',
  PRIMARY KEY (id),
  KEY idx_namespace_id (namespace_id),
  KEY idx_update_time (update_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='commit history';



# Dump of table instance
# ------------------------------------------------------------

DROP TABLE IF EXISTS instance;

CREATE TABLE instance (
  id CHAR(36) NOT NULL COMMENT '',
  app_id CHAR(32) NOT NULL COMMENT '',
  cluster_id CHAR(32) NOT NULL COMMENT '',
  host varchar(64) NOT NULL COMMENT '',
  create_time INT NOT NULL COMMENT '',
  update_time INT NULL COMMENT '',
  PRIMARY KEY (id),
  UNIQUE KEY idx_group (app_id,cluster_id,host),
  KEY idx_update_time (update_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='app instance';



# Dump of table instance_release
# ------------------------------------------------------------

DROP TABLE IF EXISTS instance_release;

CREATE TABLE instance_release (
  id INT(10) unsigned NOT NULL COMMENT '',
  instance_id CHAR(36) NOT NULL COMMENT '',
  app_id CHAR(32) NOT NULL COMMENT '',
  cluster_id CHAR(32) NOT NULL COMMENT '',
  namespace_id CHAR(32) NOT NULL  COMMENT '',
  release_id CHAR(32) NOT NULL DEFAULT '' COMMENT '',
  release_time INT NOT NULL COMMENT '',
  create_time INT NOT NULL COMMENT '',
  update_time INT NULL COMMENT '',
  PRIMARY KEY (id),
  UNIQUE KEY idx_group (instance_id,app_id,namespace_id),
  KEY idx_release_id (release_id),
  KEY idx_update_time (update_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='instance release record';



# Dump of table item
# ------------------------------------------------------------

DROP TABLE IF EXISTS item;

CREATE TABLE item (
  id CHAR(36) NOT NULL COMMENT '',
  namespace_id CHAR(36) NOT NULL COMMENT '',
  `key` VARCHAR(128) NOT NULL COMMENT 'config key',
  value LONGTEXT NOT NULL COMMENT 'config value',
  comment VARCHAR(500) DEFAULT '' COMMENT '',
  order_num INT(10) UNSIGNED DEFAULT 0 COMMENT '',
  is_delete TINYINT(1) NOT NULL DEFAULT 0 COMMENT '',
  create_by CHAR(32) NOT NULL COMMENT '',
  create_time INT NOT NULL COMMENT '',
  update_by CHAR(32) DEFAULT '' COMMENT '',
  update_time INT NULL COMMENT '',
  PRIMARY KEY (id),
  KEY idx_namespace_id (namespace_id),
  KEY idx_key (`key`),
  KEY idx_update_time (update_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='config item';



# Dump of table release
# ------------------------------------------------------------

DROP TABLE IF EXISTS `release`;

CREATE TABLE `release` (
  id CHAR(36) NOT NULL COMMENT '',
  name VARCHAR(64) NOT NULL COMMENT 'release name',
  comment VARCHAR(255) DEFAULT '' COMMENT '',
  app_id CHAR(36) NOT NULL COMMENT '',
  cluster_id CHAR(36) NOT NULL COMMENT '',
  namespace_id CHAR(36) NOT NULL COMMENT '',
  config LONGTEXT NOT NULL COMMENT '',
  is_disabled TINYINT(1) NOT NULL DEFAULT 0 COMMENT '',
  is_delete TINYINT(1) NOT NULL DEFAULT 0 COMMENT '',
  create_by CHAR(32) NOT NULL COMMENT '',
  create_time INT NOT NULL COMMENT '',
  update_by CHAR(32) DEFAULT '' COMMENT '',
  update_time INT NULL COMMENT '',
  PRIMARY KEY (id),
  KEY idx_namespace_id (namespace_id),
  KEY idx_update_time (update_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='';


# Dump of table release_history
# ------------------------------------------------------------

DROP TABLE IF EXISTS release_history;

CREATE TABLE release_history (
  id CHAR(36) NOT NULL COMMENT '',
  app_id CHAR(36) NOT NULL COMMENT '',
  cluster_id CHAR(36) NOT NULL COMMENT '',
  namespace_id CHAR(36) NOT NULL COMMENT '',
  release_id CHAR(36) NOT NULL COMMENT '',
  pre_release_id CHAR(36) NOT NULL COMMENT '',
  op_type TINYINT NOT NULL DEFAULT 0 COMMENT '0:normal 1:rollback',
  is_delete TINYINT(1) NOT NULL DEFAULT 0 COMMENT '',
  create_by CHAR(32) NOT NULL COMMENT '',
  create_time INT NOT NULL COMMENT '',
  update_by CHAR(32) DEFAULT '' COMMENT '',
  update_time INT NULL COMMENT '',
  PRIMARY KEY (id),
  KEY idx_namespace_id (namespace_id),
  KEY idx_release_id (release_id),
  KEY idx_pre_release_id (pre_release_id),
  KEY idx_update_time (update_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='';


# Dump of table setting
# ------------------------------------------------------------

DROP TABLE IF EXISTS setting;

CREATE TABLE setting (
  id CHAR(36) NOT NULL COMMENT '',
  `key` varchar(64) NOT NULL COMMENT '',
  value varchar(2048) NOT NULL COMMENT '',
  comment varchar(1024) DEFAULT '' COMMENT '',
  is_delete TINYINT(1) NOT NULL DEFAULT 0 COMMENT '',
  create_by CHAR(32) NOT NULL COMMENT '',
  create_time INT NOT NULL COMMENT '',
  update_by CHAR(32) DEFAULT '' COMMENT '',
  update_time INT NULL COMMENT '',
  PRIMARY KEY (id),
  KEY `IX_Key` (`key`),
  KEY idx_update_time (update_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='server setting';

# Config
# ------------------------------------------------------------
INSERT INTO setting (id, `key`, value, comment)
VALUES
    (uuid(), 'store.etcd.url',  'http://localhost:2379', 'etcd server url'),
    (uuid(), 'item.key.length.limit',  '128', 'item key 最大长度限制'),
    (uuid(), 'item.value.length.limit', '20000', 'item value最大长度限制');

/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
