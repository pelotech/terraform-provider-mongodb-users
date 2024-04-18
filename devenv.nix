{ pkgs, lib, config, inputs, ... }:

{

  languages.terraform = {
    enable = true;
    version = "1.8.1";
  };

  services = {
    mongodb = {
      enable = true;
      initDatabaseUsername = "root";
      initDatabasePassword = "password123";
    };
  };

  packages = [
    pkgs.go
    pkgs.terraform
    pkgs.docker-compose
    pkgs.mongosh
  ];
}
