#!/bin/bash
curl -X POST http://localhost:8080/api/v1/mail -H "Content-Type: application/json" -d @test.json