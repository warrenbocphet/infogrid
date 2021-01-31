<img src="./media/infogrid_logo.png" width="300" height="300">

# Description
A simple news aggregation. Currently support NYTimes and Reuters.

# Quick start
```cmd/main.go``` should provide a basic understanding of the package workflow

# Dependancies
| Package                           | Description                         |
|-----------------------------------|-------------------------------------|
| github.com/gorilla/mux            | routing                             |
| go.mongodb.org/mongo-driver/mongo | database (MongoDB) driver           |
| gopkg.in/jdkato/prose.v2          | text -> sentences, extract keywords |
| golang.org/x/net/html             | parsing HTML file                   |

# Usage
The application is simply a REST API.

Root URL: www.infogrid.app

```yaml
servers:
  - url: http://www.infogrid.app/
paths:
  /articles:
    get:
      summary: List all articles
      parameters:
        - name: section
          in: query
          description: section where the articles comes from (us, world, technology, etc.)
          required: false
          schema:
            type: string

        - name: tag
          in: query
          description: tag that the article contains
          required: false
          schema:
            type: string

      responses:
        '200':
          description: An array of articles
          content:
            application/json:    
              Articles:
                Article:
                  URL: string
                  Title: string
                  Section: string
                  DateCreated: "2021-01-30 00:08:43 +0000 UTC"
                  SummarisedText: string
                  Tags: list of string

  /sections:
      get:
        summary: List all available sections
  
        responses:
          '200':
            description: An array of available sections
            content:
              application/json:    
                Sections: array of string

  /tags:
      get:
        summary: List all available tags
  
        responses:
          '200':
            description: An array of available tags (biden, trump, covid-19, etc.)
            content:
              application/json:    
                Tags: array of string
```             