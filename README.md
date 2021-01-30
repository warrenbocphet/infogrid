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

# REST API
The application is simply a REST API.

Root URL: www.infogrid.app

## Show articles
* **URL**

    /articles

* **Method**

    `GET`

* **Query params (optional)**

    `section=[string]` example: us, world, etc.

    `tag=[string]` example: biden, covid-19, etc.

* **Response**

    `curl "https://www.infogrid.app/articles?section=us&tag=biden"`

```
[
    {
        "URL": "https://www.nytimes.com/2021/01/29/us/politics/biden-white-house-coronavirus.html",
        "Title": "In Biden’s White House, Masks, Closed Doors and Empty Halls",
        "Section": "us",
        "DateCreated": "2021-01-29 22:42:16 +0000 UTC",
        "SummarisedText": "WASHINGTON — Senior staff members limit interactions with each other in most offices to a total of 15 minutes in a day.",
        "Tags": [
            "white house",
            "west wing",
            "biden"
        ]
    },
    {
        "URL": "https://www.nytimes.com/2021/01/29/us/politics/biden-walter-reed-military.html",
        "Title": "Biden Visits Walter Reed, the Hospital That Treated Both Him and His Son",
        "Section": "us",
        "DateCreated": "2021-01-30 00:08:43 +0000 UTC",
        "SummarisedText": "WASHINGTON — President Biden spent six grueling months at Walter Reed National Military Medical Center more than 30 years ago, battling two brain aneurysms.",
        "Tags": [
            "biden",
            "iraq",
            "white house"
        ]
    },
    {
        "URL": "https://www.nytimes.com/2021/01/29/us/politics/biden-migrant-children-coronavirus.html",
        "Title": "Federal Court Lifts Block on Trump Policy Expelling Migrant Children at the Border",
        "Section": "us",
        "DateCreated": "2021-01-30 02:57:57 +0000 UTC",
        "SummarisedText": "WASHINGTON — A federal appeals court on Friday lifted a block on a Trump-era policy of rapidly turning away migrant children as public health risks, ramping up pressure on the Biden administration to restore the asylum process at the southwestern border.",
        "Tags": [
            "biden",
            "trump"
        ]
    },
    {
        "URL": "https://www.nytimes.com/2021/01/30/us/politics/fact-checking-biden-first-week.html",
        "Title": "Fact-Checking Biden’s First Week in Office",
        "Section": "us",
        "DateCreated": "2021-01-30 10:00:13 +0000 UTC",
        "SummarisedText": "Eleven soldiers at Fort Bliss in Texas remained hospitalized on Friday, one day after they drank antifreeze, believing it was alcohol, during a field training exercise, military officials said.",
        "Tags": [
            "biden",
            "black",
            "latino"
        ]
    }
]
```

* **Sample call**

    `$ curl "https://www.infogrid.app/articles`

    `$ curl "https://www.infogrid.app/articles?section=us"`

    `$ curl "https://www.infogrid.app/articles?tag=biden"`

    `$ curl "https://www.infogrid.app/articles?section=us&tag=biden"`

## Show available sections

* **URL**

    /sections

* **Method**

    `GET`

* **Response**

    `$ curl "https://www.infogrid.app/sections"`

    `["technology","world","us","business"]`

## Show available tags

* **URL**

    /tags

* **Method**

    `GET`

* **Response**

    `$ curl "https://www.infogrid.app/tags"`

```
["britain","jimenez","republican","armored division","tesla","latino","josh drobnyk","tibetan","grape-nuts","south carolina","south africa","gamestop","duhamel","kouchner","colonel payne","william beaumont army medical center","greene","food group","biggs","elon musk","russia","university","senate","biden","michelmore","black","tashi","british","parker","putin","musk","west wing","robinhood","portland","melbourne","trump","french","bryan","mccarthy","chinese","covid-19","iraq","gainesville","astrazeneca","white house","yellen","australian open","davis","navalny","gosar","heyman","victoria","united states","barra"]
```

