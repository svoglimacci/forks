ALTER TABLE recipes ADD CONSTRAINT recipes_cookingtime_check CHECK (cooking_time >= 0);

ALTER TABLE recipes ADD CONSTRAINT recipes_preptime_check CHECK (prep_time >= 0);

ALTER TABLE recipes ADD CONSTRAINT categories_length_check CHECK (array_length(categories, 1) BETWEEN 1 AND 5);
