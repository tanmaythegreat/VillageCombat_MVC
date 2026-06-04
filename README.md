# MVC_ASSIGNMENT
## Database
I have divided the database in 3 parts. that are 
* `JWT` that contains a user and its jwt session token data.
* `static data` this is the data that corresponds to stats of troops,buildings etc
    * defined enums are intuitive enough.
    * `building_config_base` tells the configuration of a building which is common among different types of building. Then this building is subdivided based on their categories.
    * Every building have a entries in `building_level_status` table as well the tell about the building at a particular level.
    * there is also `upgrade_costs` which tells how much money/time is required to upgrade.
    * these if `building_stats` and `building_level_stat` for each building category
    * `troops_configs` and `troop_level_stats` similar.
    * note that `upgrade_costs` takes in account for both building and troop upgrade.
    * I did this separate table level thing because earlier I used SQL array. But then I discovered that `gorm` does not support array properly. it would probably give bt in future.
* `User information`
    * `user_data` this is different from the user in first migration, it contains user game data. 
    * `placed_buildings` name is self-explanatory.
    * `trained_troops` the troops that user have and can use it for battle/match.