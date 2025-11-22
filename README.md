# rgs
Mini RGS (Remote Game Server) for the game “Lucky Dice” that can be plugged into different operators

### How to run the project:

````
1. docker compose down -v

2. docker builder prune -af

3. docker compose build --no-cache

4. docker compose up -d
````

Inside 0002_seed_data.up.sql:
```
INSERT INTO operators (name, api_key, webhook_url, webhook_secret)
VALUES
    ('Demo Operator', 'demo-operator-key', 'https://webhook.site/51380c55-679f-4468-9eb5-78235acc527e', 'demo-secret')
    ON CONFLICT DO NOTHING;
```

Change webhook URL to another for testing/production purposes.