<img src="./media/infogrid_logo.png" width="300" height="300">

# Description
A simple news aggregation. Currently support NYTimes and Reuters.

# Quick start
```cmd/main.go``` should provide a basic understanding of the package workflow

# Dependancies
For routing: ```github.com/gorilla/mux```

Official MongoDB driver: ```go.mongodb.org/mongo-driver/mongo```

Extracting keywords, might remove in the future: ```gopkg.in/jdkato/prose.v2```

For parsing HTML file: ```golang.org/x/net/html```

SMMRY, public API for text summarisation: ```https://smmry.com/```

# Road map
1. Creating a public API + website
2. Develop ~~summarisation~~ + keywords extraction in-house instead of using other framework
3. ~~Add Docker files~~
4. Deploy to AWS
5. OpenAPI compliance
