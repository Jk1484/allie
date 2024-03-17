To run the container inside the docker use:

```
docker compose up --build -d 
```

Server is available at localhost:8081



### REGISTER
```json
{"type":"register","username":"mage1","password":"mage1"}
```

### JOIN
```json
{"type":"join","username":"mage1","password":"mage1"}
```

### FIREBALL
```json
{"type":"fireball","target":"mage2"}
```
