# Wishlist service for Wili

The main purpose of this service is to encapsulate all the logic related to wishlists:

CRUD wishlist templates etc

## MVP

- Wishlist as a list of items
- Items are extensible schema-wise content pieces:
    - Just text possibly with image or link
    - Not in MVP, but need to account for it backend-wise (to support stuff like this in the future): 
        - Marketplace integration: link to an items shows as item in marketplace which you can add to the cart and buy (need only to store item id and marketplace as type: marketplace, marketplace: yandex, sku: 103538601049)
- Create wishlist, add items and display wishlist of a particular user