# Terraform Provider MongoDb-Users

### DO NOT USE THIS PROVIDER IT IS IN ACTIVE DEVELOPMENT AND MEANT TO BE CONSUMED AS A CROSSPLANE PROVIDER

A terraform provider specifically for managing users on self-hosted MongoDB instances via Crossplane

Current MongoDb providers either target a specific cloud provider offering (Atlas, DocumentDb, CosmosDb) or are incompatible with the requirements for conversion into a crossplane provider.


#### Development

To open a shell with the project dependencies:

```shell
env NIXPKGS_ALLOW_UNFREE=1 devenv --impure shell
```
-


To run a mongodb instance to test this project against:

| :warning: WARNING          |
|:---------------------------|
| THIS BUILDS A MONGODB INSTANCE FROM SOURCE AND CAN TAKE 40+ MINS ON FIRST RUN      |


```shell
env NIXPKGS_ALLOW_UNFREE=1 devenv --impure up
```

Alternatively there is a docker-compose in `docker-compose/docker-compose.yml`

```shell
docker compose up
```
--

Once there is a running mongodb instance you can run the acceptance tests with:

```shell
TF_ACC=1 go test ./...
```

