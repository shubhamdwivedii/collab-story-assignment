USE collab; 

CREATE TABLE stories (
    id int NOT NULL AUTO_INCREMENT,
    title varchar(22) DEFAULT '' NOT NULL, 
    titleAdded tinyint(1) DEFAULT 0,
    isFinished tinyint(1) DEFAULT 0, 
    createdAt datetime DEFAULT CURRENT_TIMESTAMP NOT NULL, 
    updatedAt datetime DEFAULT CURRENT_TIMESTAMP NOT NULL, 
    PRIMARY KEY (id)
)

CREATE TABLE paragraphs (
    id int NOT NULL AUTO_INCREMENT, 
    story int NOT NULL, 
    isFinished tinyint(1) DEFAULT 0, 
    PRIMARY KEY (id)
)

CREATE TABLE sentences (
    id int NOT NULL AUTO_INCREMENT, 
    paragraph int NOT NULL, 
    isFinished tinyint(1) DEFAULT 0, 
    content varchar(240) DEFAULT '', -- 15 chars in 15 words + 14 spaces in between ~ 240
    PRIMARY KEY (id)
)