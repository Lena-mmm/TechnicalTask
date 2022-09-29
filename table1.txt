CREATE TABLE public."Table1"
(
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 ),
    link text NOT NULL,
    short character varying(40) NOT NULL,
    PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public."Table1"
    OWNER to postgres;