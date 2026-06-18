# MVC_ASSIGNMENT
## Database
I have divided the database in 4 parts. that are 
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
* `Battle data` when 2 players battles .i.e one player attacks the other then 
    * I store the outcome of the battle that is what the attacker got (looted).
    * also storing data about looses by both sides in separate tables

## JSON WEB TOKENS
I have implemented jwt with rotating refresh tokens.
`AccessToken` have 3 parts `header.payload.signature` `signature = sha256(header.payload,key)`. I also generate a Random some bytes using `crypto/rand` (and not `math/rand`) which is the refresh token. when the token is about to expire frontend will ask the backend to refresh the `AccessToken` and this `RefreshToken` acts as password. I store the bcrypt hash of `RefreshToken` in the database.

register and login handler are straightforward. at the end it generates JWT token.
I tried my best to avoid time based vulnerabilities.

## Frontend
I am using Three.js to render the 3d environment
After signing in the tokens are stored in localstorage,then it Redirects to game page, when the game loads it reads the token from local storage again
Then it connects to the server via web socket,on successful Connection it asks the server to send the placed building data,and then asks for all building data if all building data is not in the local storage.
when both building data are ready I load the map.
* `Building Spawning Algorithm`
  * Before loading objects I first make all the loaded objects invisible.and add them to the pool. 
  * while spawning i check if there is a pooled object that can be moved there I move it and make it visible.(spawning objects is heavier than moving)
  * if there is no pooled object then I spawn.

First of all you will see a building (ignore the model for now) that's town hall.
you can only create `Cannon` and `Town Hall` for now.
For now construction is instantaneous (construction task is not involved).
There is no interactive way(for now).
in the console write `CreateBuilding('Cannon',30,20)`

the Above for now is not now.

spawn building by clicking at blank surface
open building menu to upgrade and all 

Issues : move is not working , you can create another townhall (may be i will keep it as a feature).
TODO : add a mechanism to collect resources from gold mines , elixir drill , elixir collect.Troops training.Battle.

i tried to find animated 3d models of troops but couldn't so images moves.
during the battle access token is to checked (will fix)

there are a bunch of TODO all over the place, which i will do , the bare minimum prototype is ready

issues:
there are some bugs related to positioning of buildings
broken buildings are not shown (it appears normal).
there are some bugs in the battle.

repairCost formulas : upgradeCost/10,for gems (1+updrageCost/10) 
i randomly came up with it, no back story.

repair is working 
also now attackers can also attack defenders
grid highlight adds even more feel