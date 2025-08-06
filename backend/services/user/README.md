# User service for Wili

The main purpose of this service is to encapsulate all the logic related to users:

auth/profiles/etc.

## MVP

- Yandex ID signup/singin (for now, no from-scratch accounts, only via other providers, but store them internally)
- Simple profiles: name, pic (which you can grab from Yandex ID)
- OAuth to interact with wishlist service & validate auth from wishlist service
- Rest API