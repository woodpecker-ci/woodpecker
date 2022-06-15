# Architecture

## Package architecture

![Woodpecker architecture](./woodpecker-architecture.png)

## System architecture

### Server

```none
server/router -> server/api -> server/pipeline -> pipeline/*
                                               -> server/store
```
