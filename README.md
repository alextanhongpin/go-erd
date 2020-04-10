# go-erd

Converts this text into an Entity Relationship Diagram:

```
[Person] {color: "red"}
*name
height
weight
birth_date
+birth_place_id

[Birth Place]
*id
birth_city
birth state
birth country

[Roles]
*id
name
+user_id

# Each relationship must be between exactly two entities, which need not
# be distinct. Each entity in the relationship has exactly one of four
# possible cardinalities:
#
# Cardinality    Syntax
# 0 or 1         ?
# exactly 1      1
# 0 or more      *
# 1 or more      +
Person *--1 Birth Place
Person *--1 Roles
```

Output:

![out.png](./assets/out.png)
