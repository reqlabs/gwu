# Poem API Example

This directory contains a simple in-memory Poem Store with a JSON API to demonstrate the usage of the Gwu web utility package.

## Summary

The Poem API allows you to create and retrieve poems using HTTP requests. This example demonstrates how to set up and use Gwu in a practical application.

## Routes

- **Get Poem By ID**
    - **Method:** GET
    - **URL:** `/poem/{id}`
    - **Description:** Retrieves a poem by its ID.
    - **Example:** `localhost:8080/poem/isjzB57elf`

- **Get Poems By Author**
    - **Method:** GET
    - **URL:** `/poems/author/{author}`
    - **Description:** Retrieves poems by a specific author.
    - **Example:** `localhost:8080/poems/author/Goethe`

- **Get All Poems**
    - **Method:** GET
    - **URL:** `/poems`
    - **Description:** Retrieves all poems.
    - **Example:** `localhost:8080/poems`

- **Create Poem**
    - **Method:** POST
    - **URL:** `/poem`
    - **Description:** Creates a new poem.
    - **Example:**
        - **URL:** `localhost:8080/poem`
        - **Body:**
      ```json
      {
          "name": "Der Zauberlehrling",
          "author": "Goethe",
          "text": "Hat der alte Hexenmeister\nsich doch einmal wegbegeben!\nUnd nun sollen seine Geister\nauch nach meinem Willen leben.\nSeine Wort' und Werke\nMerkt ich und den Brauch,\nund mit Geistesstärke\ntu ich Wunder auch."
      }
      ```

## How to Run

You can run the application by executing the following command from the root directory:

```sh
go run ./examples/poem
```

## Postman Collection

To try out the API using Postman, use the provided [Postman collection](gwu_poem_example.postman_collection.json).