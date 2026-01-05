package confy

import (
	"testing"
)

func TestWriteConfig(t *testing.T) {
	SetConfigFile("config.yml")
	ReadInConfig()
	Set("mysql.path", "127.0.0.1")
	Set("mysql.port", "3306")
	Set("mysql.config", "charset=utf8mb4&parseTime=True&loc=Local")
	Set("mysql.db-name", "")
	Set("mysql.username", "test")
	Set("mysql.password", "test")
	Set("mysql.prefix", "")
	Set("mysql.singular", false)
	Set("mysql.engine", "")
	Set("mysql.max-idle-conns", 10)
	Set("mysql.max-open-conns", 100)
	Set("mysql.log-mode", "error")
	Set("jwt.signing-key", "131a5a9e-ccf4-434f-b17c-ed46bda2c4da")
	err := WriteConfig()
	if err != nil {
		t.Errorf("write config failed:%v", err)
	}
}
