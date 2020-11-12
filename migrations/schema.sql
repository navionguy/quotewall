--
-- PostgreSQL database dump
--

-- Dumped from database version 10.13
-- Dumped by pg_dump version 10.13

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: pick_from_range(integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.pick_from_range(bottom integer, top integer) RETURNS integer
    LANGUAGE plpgsql STRICT
    AS $$
BEGIN
   RETURN FLOOR(random()* (top-bottom + 1) + bottom);
END;
$$;


ALTER FUNCTION public.pick_from_range(bottom integer, top integer) OWNER TO postgres;

--
-- Name: shuffle_deck(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.shuffle_deck() RETURNS integer
    LANGUAGE plpgsql STRICT
    AS $$
DECLARE
    max_rec     integer;
    i           integer;
    j           integer;
    keys        uuid[];
    marker      text;
BEGIN

    /* fastest way to clear the table */
    IF EXISTS (SELECT * FROM pg_tables WHERE tablename='shuffled_conversations')
         THEN
             DROP TABLE shuffled_conversations;
    END IF;    

    CREATE TABLE shuffled_conversations (
        id              uuid NOT NULL,
        sequence        integer NOT NULL PRIMARY KEY
    );
    ALTER TABLE shuffled_conversations
        ADD CONSTRAINT id_fkey FOREIGN KEY (id) REFERENCES public.conversations(id) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;

    keys := ARRAY(SELECT id FROM conversations
                        WHERE publish = TRUE);    /* load up all the published conversation ID values */
    i := 0;                                       /* rolls over the entire array doing the shuffle */
    max_rec := array_length(keys,1);              /* get number of conversations in the array */

    LOOP
        i := i + 1; /* move forward, there is no 0 element */

        /* pick a random element still in the array */
        /* insert it into the current position */
        /* then put the current element into its position in the array */
        /* by the time I'm done, the Keys array is trashed, don't try to use it */

        j := pick_from_range(i,max_rec);    
        INSERT INTO shuffled_conversations( sequence, ID) VALUES( i, keys[j] );
        keys[j] := keys[i];

        EXIT WHEN i = max_rec;
    END LOOP;

    /* set the current date as a comment on the table */
    marker := (SELECT CURRENT_DATE);
    EXECUTE FORMAT('COMMENT ON TABLE shuffled_conversations IS ''%I''', marker);

    /* and the record count as a comment on the id column */
    EXECUTE FORMAT('COMMENT ON COLUMN shuffled_conversations.sequence IS ''%I''', max_rec);
    
    /* tag this run of the record shuffle */
    keys[1] := (SELECT uuid_generate_v4());
    EXECUTE FORMAT('COMMENT ON COLUMN shuffled_conversations.id IS ''%I''', keys[1]);
    
    RETURN max_rec;
END
$$;


ALTER FUNCTION public.shuffle_deck() OWNER TO postgres;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: annotations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.annotations (
    id uuid NOT NULL,
    note character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.annotations OWNER TO postgres;

--
-- Name: author_counts; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.author_counts AS
SELECT
    NULL::uuid AS id,
    NULL::character varying(255) AS name,
    NULL::timestamp without time zone AS created_at,
    NULL::timestamp without time zone AS updated_at,
    NULL::bigint AS count;


ALTER TABLE public.author_counts OWNER TO postgres;

--
-- Name: authors; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.authors (
    id uuid NOT NULL,
    name character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.authors OWNER TO postgres;

--
-- Name: conversations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.conversations (
    id uuid NOT NULL,
    occurredon timestamp without time zone NOT NULL,
    publish boolean NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.conversations OWNER TO postgres;

--
-- Name: permissions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.permissions (
    id uuid NOT NULL,
    name character varying(255) NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.permissions OWNER TO postgres;

--
-- Name: quotes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.quotes (
    id uuid NOT NULL,
    saidon timestamp without time zone NOT NULL,
    sequence integer NOT NULL,
    phrase character varying(255) NOT NULL,
    publish boolean NOT NULL,
    annotation_id uuid,
    author_id uuid NOT NULL,
    conversation_id uuid NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.quotes OWNER TO postgres;

--
-- Name: schema_migration; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.schema_migration (
    version character varying(14) NOT NULL
);


ALTER TABLE public.schema_migration OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: annotations annotations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.annotations
    ADD CONSTRAINT annotations_pkey PRIMARY KEY (id);


--
-- Name: authors authors_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.authors
    ADD CONSTRAINT authors_pkey PRIMARY KEY (id);


--
-- Name: conversations conversations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: quotes quotes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quotes
    ADD CONSTRAINT quotes_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: schema_migration_version_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX schema_migration_version_idx ON public.schema_migration USING btree (version);


--
-- Name: author_counts _RETURN; Type: RULE; Schema: public; Owner: postgres
--

CREATE OR REPLACE VIEW public.author_counts AS
 SELECT a.id,
    a.name,
    a.created_at,
    a.updated_at,
    count(a.id) AS count
   FROM (public.authors a
     JOIN public.quotes q ON ((a.id = q.author_id)))
  GROUP BY a.id;


--
-- Name: permissions permissions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;


--
-- Name: quotes quotes_annotation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quotes
    ADD CONSTRAINT quotes_annotation_id_fkey FOREIGN KEY (annotation_id) REFERENCES public.annotations(id);


--
-- Name: quotes quotes_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quotes
    ADD CONSTRAINT quotes_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.authors(id) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;


--
-- Name: quotes quotes_conversation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quotes
    ADD CONSTRAINT quotes_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES public.conversations(id) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;


--
-- PostgreSQL database dump complete
--

