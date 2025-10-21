--
-- PostgreSQL database dump
--

\restrict 18NccrfpihYnwqBivs51uHtRdmaM3893BGxRvmp6T7W9a7eNFNYZo0q0ioWALlU

-- Dumped from database version 17.6 (Debian 17.6-2.pgdg13+1)
-- Dumped by pg_dump version 17.6

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: agents; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.agents (
    id bigint NOT NULL,
    created bigint,
    updated bigint,
    name character varying(255),
    owner_id bigint,
    token character varying(255),
    last_contact bigint,
    platform character varying(100),
    backend character varying(100),
    capacity integer,
    version character varying(255),
    no_schedule boolean,
    last_work bigint,
    org_id bigint,
    custom_labels json
);


ALTER TABLE public.agents OWNER TO postgres;

--
-- Name: agents_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.agents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.agents_id_seq OWNER TO postgres;

--
-- Name: agents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.agents_id_seq OWNED BY public.agents.id;


--
-- Name: pipelines; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pipelines (
    id integer NOT NULL,
    repo_id integer,
    number integer,
    event character varying(500),
    status character varying(500),
    created integer,
    started integer,
    finished integer,
    commit character varying(500),
    branch character varying(500),
    ref character varying(500),
    refspec character varying(1000),
    title character varying(1000),
    message text,
    "timestamp" integer,
    author character varying(500),
    avatar character varying(1000),
    email character varying(500),
    forge_url character varying(1000),
    deploy character varying(500),
    parent integer,
    reviewer character varying(250),
    reviewed integer,
    sender character varying(250),
    changed_files text,
    updated bigint DEFAULT 0 NOT NULL,
    additional_variables json,
    pr_labels json,
    errors json,
    deploy_task character varying(255),
    is_prerelease boolean,
    from_fork boolean
);


ALTER TABLE public.pipelines OWNER TO postgres;

--
-- Name: builds_build_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.builds_build_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.builds_build_id_seq OWNER TO postgres;

--
-- Name: builds_build_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.builds_build_id_seq OWNED BY public.pipelines.id;


--
-- Name: configs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.configs (
    id integer NOT NULL,
    repo_id integer,
    hash character varying(250),
    data bytea,
    name text
);


ALTER TABLE public.configs OWNER TO postgres;

--
-- Name: config_config_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.config_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.config_config_id_seq OWNER TO postgres;

--
-- Name: config_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.config_config_id_seq OWNED BY public.configs.id;


--
-- Name: crons; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.crons (
    id bigint NOT NULL,
    name character varying(255),
    repo_id bigint,
    creator_id bigint,
    next_exec bigint,
    schedule character varying(255) NOT NULL,
    created bigint DEFAULT 0 NOT NULL,
    branch character varying(255)
);


ALTER TABLE public.crons OWNER TO postgres;

--
-- Name: crons_i_d_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.crons_i_d_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.crons_i_d_seq OWNER TO postgres;

--
-- Name: crons_i_d_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.crons_i_d_seq OWNED BY public.crons.id;


--
-- Name: forges; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.forges (
    id bigint NOT NULL,
    type character varying(250),
    url character varying(500),
    client character varying(250),
    client_secret character varying(250),
    skip_verify boolean,
    oauth_host character varying(250),
    additional_options json
);


ALTER TABLE public.forges OWNER TO postgres;

--
-- Name: forge_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.forge_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.forge_id_seq OWNER TO postgres;

--
-- Name: forge_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.forge_id_seq OWNED BY public.forges.id;


--
-- Name: log_entries; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.log_entries (
    id bigint NOT NULL,
    step_id bigint,
    "time" bigint,
    line integer,
    data bytea,
    created bigint,
    type integer
);


ALTER TABLE public.log_entries OWNER TO postgres;

--
-- Name: log_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.log_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.log_entries_id_seq OWNER TO postgres;

--
-- Name: log_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.log_entries_id_seq OWNED BY public.log_entries.id;


--
-- Name: migration; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.migration (
    id character varying(255),
    description character varying(255)
);


ALTER TABLE public.migration OWNER TO postgres;

--
-- Name: orgs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.orgs (
    id bigint NOT NULL,
    name character varying(255),
    is_user boolean,
    private boolean,
    forge_id bigint
);


ALTER TABLE public.orgs OWNER TO postgres;

--
-- Name: orgs_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.orgs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.orgs_id_seq OWNER TO postgres;

--
-- Name: orgs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.orgs_id_seq OWNED BY public.orgs.id;


--
-- Name: perms; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.perms (
    user_id integer NOT NULL,
    repo_id integer NOT NULL,
    pull boolean,
    push boolean,
    admin boolean,
    synced integer,
    created bigint,
    updated bigint
);


ALTER TABLE public.perms OWNER TO postgres;

--
-- Name: pipeline_configs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pipeline_configs (
    config_id bigint NOT NULL,
    pipeline_id bigint NOT NULL
);


ALTER TABLE public.pipeline_configs OWNER TO postgres;

--
-- Name: steps; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.steps (
    id integer NOT NULL,
    pipeline_id integer,
    pid integer,
    ppid integer,
    name character varying(250),
    state character varying(250),
    error text,
    exit_code integer,
    started integer,
    finished integer,
    uuid character varying(255),
    failure character varying(255),
    type character varying(255)
);


ALTER TABLE public.steps OWNER TO postgres;

--
-- Name: procs_proc_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.procs_proc_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.procs_proc_id_seq OWNER TO postgres;

--
-- Name: procs_proc_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.procs_proc_id_seq OWNED BY public.steps.id;


--
-- Name: redirections; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.redirections (
    id bigint NOT NULL,
    repo_id bigint,
    repo_full_name character varying(255)
);


ALTER TABLE public.redirections OWNER TO postgres;

--
-- Name: redirections_redirection_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.redirections_redirection_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.redirections_redirection_id_seq OWNER TO postgres;

--
-- Name: redirections_redirection_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.redirections_redirection_id_seq OWNED BY public.redirections.id;


--
-- Name: registries; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.registries (
    id integer NOT NULL,
    repo_id integer DEFAULT 0 NOT NULL,
    address character varying(250) NOT NULL,
    username character varying(2000),
    password text,
    org_id bigint DEFAULT 0 NOT NULL
);


ALTER TABLE public.registries OWNER TO postgres;

--
-- Name: registry_registry_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.registry_registry_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.registry_registry_id_seq OWNER TO postgres;

--
-- Name: registry_registry_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.registry_registry_id_seq OWNED BY public.registries.id;


--
-- Name: repos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.repos (
    id integer NOT NULL,
    user_id integer,
    owner character varying(250),
    name character varying(250),
    full_name character varying(250),
    avatar character varying(500),
    forge_url character varying(1000),
    clone character varying(1000),
    branch character varying(500),
    timeout integer,
    private boolean,
    allow_pr boolean,
    repo_allow_push boolean,
    hash character varying(500),
    config_path character varying(500),
    visibility character varying(50),
    active boolean,
    forge_remote_id character varying(255),
    org_id bigint,
    cancel_previous_pipeline_events json,
    clone_ssh character varying(1000),
    pr_enabled boolean DEFAULT true,
    forge_id bigint,
    allow_deploy boolean,
    require_approval character varying(255),
    trusted json,
    netrc_trusted json
);


ALTER TABLE public.repos OWNER TO postgres;

--
-- Name: repos_repo_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.repos_repo_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.repos_repo_id_seq OWNER TO postgres;

--
-- Name: repos_repo_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.repos_repo_id_seq OWNED BY public.repos.id;


--
-- Name: secrets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.secrets (
    id integer NOT NULL,
    repo_id integer DEFAULT 0 NOT NULL,
    name character varying(250) NOT NULL,
    value bytea,
    images character varying(2000),
    events character varying(2000),
    org_id bigint DEFAULT 0 NOT NULL
);


ALTER TABLE public.secrets OWNER TO postgres;

--
-- Name: secrets_secret_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.secrets_secret_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.secrets_secret_id_seq OWNER TO postgres;

--
-- Name: secrets_secret_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.secrets_secret_id_seq OWNED BY public.secrets.id;


--
-- Name: server_configs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.server_configs (
    key character varying(255) NOT NULL,
    value character varying(255)
);


ALTER TABLE public.server_configs OWNER TO postgres;

--
-- Name: tasks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tasks (
    id character varying(250) NOT NULL,
    data bytea,
    labels bytea,
    dependencies bytea,
    run_on bytea,
    dependencies_status json,
    agent_id bigint
);


ALTER TABLE public.tasks OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer NOT NULL,
    login character varying(250),
    access_token text,
    refresh_token text,
    expiry integer,
    email character varying(500),
    avatar character varying(500),
    admin boolean,
    hash character varying(500),
    forge_remote_id character varying(255),
    org_id bigint,
    forge_id bigint
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_user_id_seq OWNER TO postgres;

--
-- Name: users_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_user_id_seq OWNED BY public.users.id;


--
-- Name: workflows; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.workflows (
    id bigint NOT NULL,
    pipeline_id bigint,
    pid integer,
    name character varying(255),
    state character varying(255),
    error text,
    started bigint,
    finished bigint,
    agent_id bigint,
    platform character varying(255),
    environ json,
    axis_id integer
);


ALTER TABLE public.workflows OWNER TO postgres;

--
-- Name: workflows_workflow_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.workflows_workflow_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.workflows_workflow_id_seq OWNER TO postgres;

--
-- Name: workflows_workflow_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.workflows_workflow_id_seq OWNED BY public.workflows.id;


--
-- Name: agents id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.agents ALTER COLUMN id SET DEFAULT nextval('public.agents_id_seq'::regclass);


--
-- Name: configs id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.configs ALTER COLUMN id SET DEFAULT nextval('public.config_config_id_seq'::regclass);


--
-- Name: crons id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.crons ALTER COLUMN id SET DEFAULT nextval('public.crons_i_d_seq'::regclass);


--
-- Name: forges id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forges ALTER COLUMN id SET DEFAULT nextval('public.forge_id_seq'::regclass);


--
-- Name: log_entries id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.log_entries ALTER COLUMN id SET DEFAULT nextval('public.log_entries_id_seq'::regclass);


--
-- Name: orgs id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orgs ALTER COLUMN id SET DEFAULT nextval('public.orgs_id_seq'::regclass);


--
-- Name: pipelines id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipelines ALTER COLUMN id SET DEFAULT nextval('public.builds_build_id_seq'::regclass);


--
-- Name: redirections id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.redirections ALTER COLUMN id SET DEFAULT nextval('public.redirections_redirection_id_seq'::regclass);


--
-- Name: registries id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.registries ALTER COLUMN id SET DEFAULT nextval('public.registry_registry_id_seq'::regclass);


--
-- Name: repos id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repos ALTER COLUMN id SET DEFAULT nextval('public.repos_repo_id_seq'::regclass);


--
-- Name: secrets id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.secrets ALTER COLUMN id SET DEFAULT nextval('public.secrets_secret_id_seq'::regclass);


--
-- Name: steps id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.steps ALTER COLUMN id SET DEFAULT nextval('public.procs_proc_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_user_id_seq'::regclass);


--
-- Name: workflows id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.workflows ALTER COLUMN id SET DEFAULT nextval('public.workflows_workflow_id_seq'::regclass);


--
-- Data for Name: agents; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.agents (id, created, updated, name, owner_id, token, last_contact, platform, backend, capacity, version, no_schedule, last_work, org_id, custom_labels) FROM stdin;
1	1641630000	1641630000	agent-1	1	agent_token_abc123xyz	1641630000	linux	docker	2	1.0.0	f	\N	-1	\N
2	1641630100	1641630100	agent-2	1	agent_token_def456uvw	1641630100	linux	docker	4	1.0.0	f	\N	-1	\N
3	1641630200	1641630200	agent-3	2	agent_token_ghi789rst	1641630200	linux	kubernetes	8	1.0.1	f	\N	-1	\N
\.


--
-- Data for Name: configs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.configs (id, repo_id, hash, data, name) FROM stdin;
1	105	ec8ca9529d6081e631aec26175b26ac91699395b96b9c5fc1f3af6d3aef5d3a8	\\x636c6f6e653a0a20206769743a0a20202020696d6167653a20776f6f647065636b657263692f706c7567696e2d6769743a746573740a0a706970656c696e653a0a20205072696e743a0a20202020696d6167653a207072696e742f656e760a20202020736563726574733a205b204141414141414141414141414141414141414141414141414141205d	drone
\.


--
-- Data for Name: crons; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.crons (id, name, repo_id, creator_id, next_exec, schedule, created, branch) FROM stdin;
1	nightly-build	105	1	1641686400	0 0 * * *	1641630600	master
\.


--
-- Data for Name: forges; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.forges (id, type, url, client, client_secret, skip_verify, oauth_host, additional_options) FROM stdin;
1	gitea	http://100.114.106.50:3000	6e9119df-a86d-4fe0-b392-fe125d7a265f	gto_bagkxxp5yio7npmj7uzrf5neyyalfbqykfmri3ryqfpgvlylqwsa	f		{}
\.


--
-- Data for Name: log_entries; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.log_entries (id, step_id, "time", line, data, created, type) FROM stdin;
1	2	0	0	\\x537465704e616d653a20636c6f6e65	1641630525	0
2	2	0	1	\\x53746570547970653a20636c6f6e65	1641630525	0
3	2	0	2	\\x53746570555549443a2030314a3151344e443232594b534a31465a443654533234343357	1641630525	0
4	2	0	3	\\x53746570436f6d6d616e64733a	1641630525	0
5	2	0	4	\\x2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d	1641630525	0
6	2	0	5	\\x	1641630525	0
7	2	0	6	\\x2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d	1641630525	0
8	2	0	7	\\x	1641630525	0
9	3	0	0	\\x537465704e616d653a205072696e74	1641630526	0
10	3	0	1	\\x53746570547970653a20636f6d6d616e6473	1641630526	0
11	3	0	2	\\x53746570555549443a2030314a3151344e443232594b534a31465a44365739385a573047	1641630526	0
12	3	0	3	\\x53746570436f6d6d616e64733a	1641630526	0
13	3	0	4	\\x2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d	1641630526	0
14	3	0	5	\\x7072696e7420656e7620636f6d6d616e64	1641630526	0
15	3	0	6	\\x2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d	1641630526	0
16	3	0	7	\\x	1641630526	0
\.


--
-- Data for Name: migration; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.migration (id, description) FROM stdin;
SCHEMA_INIT	
legacy-to-xormigrate	
add-org-id	
alter-table-tasks-update-type-of-task-data	
alter-table-config-update-type-of-config-data	
remove-plugin-only-option-from-secrets-table	
convert-to-new-pipeline-error-format	
rename-link-to-url	
clean-registry-pipeline	
set-forge-id	
unify-columns-tables	
alter-table-registries-fix-required-fields	
correct-potential-corrupt-orgs-users-relation	
gated-to-require-approval	
cron-without-sec	
rename-start-end-time	
fix-v31-registries	
remove-old-migrations-of-v1	
add-org-agents	
add-custom-labels-to-agent	
split-trusted	
remove-repo-netrc-only-trusted	
rename-token-fields	
set-new-defaults-for-require-approval	
remove-repo-scm	
\.


--
-- Data for Name: orgs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.orgs (id, name, is_user, private, forge_id) FROM stdin;
1	2	f	f	1
2	test	t	f	1
\.


--
-- Data for Name: perms; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.perms (user_id, repo_id, pull, push, admin, synced, created, updated) FROM stdin;
1	1	t	t	t	1641626844	\N	\N
1	2	t	t	t	1641626844	\N	\N
1	3	t	t	t	1641626844	\N	\N
1	4	t	t	t	1641626844	\N	\N
1	5	t	t	t	1641626844	\N	\N
1	6	t	t	t	1641626844	\N	\N
1	7	t	t	t	1641626844	\N	\N
1	8	t	t	t	1641626844	\N	\N
1	9	t	t	t	1641626844	\N	\N
1	10	t	t	t	1641626844	\N	\N
1	11	t	t	t	1641626844	\N	\N
1	12	t	t	t	1641626844	\N	\N
1	13	t	t	t	1641626844	\N	\N
1	14	t	t	t	1641626844	\N	\N
1	15	t	t	t	1641626844	\N	\N
1	16	t	t	t	1641626844	\N	\N
1	17	t	t	t	1641626844	\N	\N
1	18	t	t	t	1641626844	\N	\N
1	19	t	t	t	1641626844	\N	\N
1	20	t	t	t	1641626844	\N	\N
1	21	t	t	t	1641626844	\N	\N
1	22	t	t	t	1641626844	\N	\N
1	23	t	t	t	1641626844	\N	\N
1	24	t	t	t	1641626844	\N	\N
1	25	t	t	t	1641626844	\N	\N
1	26	t	t	t	1641626844	\N	\N
1	27	t	t	t	1641626844	\N	\N
1	28	t	t	t	1641626844	\N	\N
1	29	t	t	t	1641626844	\N	\N
1	30	t	t	t	1641626844	\N	\N
1	31	t	t	t	1641626844	\N	\N
1	32	t	t	t	1641626844	\N	\N
1	33	t	t	t	1641626844	\N	\N
1	34	t	t	t	1641626844	\N	\N
1	35	t	t	t	1641626844	\N	\N
1	36	t	t	t	1641626844	\N	\N
1	37	t	t	t	1641626844	\N	\N
1	38	t	t	t	1641626844	\N	\N
1	39	t	t	t	1641626844	\N	\N
1	40	t	t	t	1641626844	\N	\N
1	41	t	t	t	1641626844	\N	\N
1	42	t	t	t	1641626844	\N	\N
1	43	t	t	t	1641626844	\N	\N
1	44	t	t	t	1641626844	\N	\N
1	45	t	t	t	1641626844	\N	\N
1	46	t	t	t	1641626844	\N	\N
1	47	t	t	t	1641626844	\N	\N
1	48	t	t	t	1641626844	\N	\N
1	49	t	t	t	1641626844	\N	\N
1	50	t	t	t	1641626844	\N	\N
1	51	t	t	t	1641626844	\N	\N
1	52	t	t	t	1641626844	\N	\N
1	53	t	t	t	1641626844	\N	\N
1	54	t	t	t	1641626844	\N	\N
1	55	t	t	t	1641626844	\N	\N
1	56	t	t	t	1641626844	\N	\N
1	57	t	t	t	1641626844	\N	\N
1	58	t	t	t	1641626844	\N	\N
1	59	t	t	t	1641626844	\N	\N
1	60	t	t	t	1641626844	\N	\N
1	115	t	t	t	1641630451	\N	\N
1	105	t	t	t	1641630452	\N	\N
\.


--
-- Data for Name: pipeline_configs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pipeline_configs (config_id, pipeline_id) FROM stdin;
1	1
\.


--
-- Data for Name: pipelines; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pipelines (id, repo_id, number, event, status, created, started, finished, commit, branch, ref, refspec, title, message, "timestamp", author, avatar, email, forge_url, deploy, parent, reviewer, reviewed, sender, changed_files, updated, additional_variables, pr_labels, errors, deploy_task, is_prerelease, from_fork) FROM stdin;
1	105	1	push	failure	1641630525	1641630525	1641630527	24bf205107cea48b92bc6444e18e40d21733a594	master	refs/heads/master			„.drone.yml“ hinzufügen\\n	1641630525	test	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	test@test.test	http://10.40.8.5:3000/2/settings/compare/3fee083df05667d525878b5fcbd4eaf2a121c559...24bf205107cea48b92bc6444e18e40d21733a594		0		0	test	[".drone.yml"]\\n	0	\N	\N	\N	\N	\N	\N
\.


--
-- Data for Name: redirections; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.redirections (id, repo_id, repo_full_name) FROM stdin;
\.


--
-- Data for Name: registries; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.registries (id, repo_id, address, username, password, org_id) FROM stdin;
\.


--
-- Data for Name: repos; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.repos (id, user_id, owner, name, full_name, avatar, forge_url, clone, branch, timeout, private, allow_pr, repo_allow_push, hash, config_path, visibility, active, forge_remote_id, org_id, cancel_previous_pipeline_events, clone_ssh, pr_enabled, forge_id, allow_deploy, require_approval, trusted, netrc_trusted) FROM stdin;
115	1	2	testCIservices	2/testCIservices	http://10.40.8.5:3000/avatars/c81e728d9d4c2f636f067f89cc14862c	http://10.40.8.5:3000/2/testCIservices	http://10.40.8.5:3000/2/testCIservices.git	master	60	f	t	t	FOUXTSNL2GXK7JP2SQQJVWVAS6J4E4SGIQYPAHEJBIFPVR46LLDA====	.drone.yml	public	t	\N	1	\N	\N	t	1	\N	forks	{"network":false,"volumes":false,"security":false}	\N
105	1	2	settings	2/settings	http://10.40.8.5:3000/avatars/c81e728d9d4c2f636f067f89cc14862c	http://10.40.8.5:3000/2/settings	http://10.40.8.5:3000/2/settings.git	master	60	f	t	t	3OQA7X5CNGPTILDYLQSJFDML6U2W7UUFBPPP2G2LRBG3WETAYZLA====	.drone.yml	public	t	\N	1	\N	\N	t	1	\N	forks	{"network":false,"volumes":false,"security":false}	\N
\.


--
-- Data for Name: secrets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.secrets (id, repo_id, name, value, images, events, org_id) FROM stdin;
1	105	wow	\\x74657374	null\\n	["push","tag","deployment","pull_request"]\\n	0
2	105	n	\\x6e	null\\n	["deployment"]\\n	0
3	105	abc	\\x656466	null\\n	["push"]\\n	0
4	105	quak	\\x66647361	null\\n	["pull_request"]\\n	0
\.


--
-- Data for Name: server_configs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.server_configs (key, value) FROM stdin;
signature-private-key	1fe3b71c87d7f89fa878306028cf08d66020ef6cafc2af90d05c40ebd03eee3c93189d2a3c46fe5292afc33e9237615ed595ee3d588dce431d5f6848b6a9bf77
jwt-secret	GKQDHRJXNN5ONIMOHJUMYDBR4IYIH46M6E5HOXX3Q2KEVZ35GM5Q====
\.


--
-- Data for Name: steps; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.steps (id, pipeline_id, pid, ppid, name, state, error, exit_code, started, finished, uuid, failure, type) FROM stdin;
2	1	2	1	git	success		0	1641630525	1641630527	\N	\N	\N
3	1	3	1	Print	skipped		0	0	0	\N	\N	\N
\.


--
-- Data for Name: tasks; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tasks (id, data, labels, dependencies, run_on, dependencies_status, agent_id) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, login, access_token, refresh_token, expiry, email, avatar, admin, hash, forge_remote_id, org_id, forge_id) FROM stdin;
1	test	eyJhbGciOiJSUzI1NiIsImtpZCI6IldmbUJ1c2Q0RndUVWRmMjc2NHowUWlEYlJ3TnRBcU5pNVlXS1U1c2k0eEEiLCJ0eXAiOiJKV1QifQ.eyJnbnQiOjEsInR0IjowLCJleHAiOjE2NDE2MzQxMjcsImlhdCI6MTY0MTYzMDUyN30.Fu0wUP-08NpPjq737y6HOeyKN_-_SE4iOZr5yrH7S8Jrz8nIuNKfU7AvlypeMSJ7wo8e3cSTadbSH1polZuFv-Nb1AqWDDXeuXudm61BkF96sTslbSHd0nF7cOy6hqCfIAfQLQpqZTJZ4E26oOSSJxPfOOntOWhlEejRl5F-flXAoYAQLegHxdn9IfYJeM1eanZqF4k6dT9hthFp9v4fmUjODPPfHip_iS7ckPonP1E4-8KeNkU3O-lIS1fgrsbCDA8531FXIGB0U7cSur7H0picKGL6WSzAErPGntlNlQWYB5JedDtLN9Ionxy1Y9LKQON76XYL4gM1Ji98RCEXggVqd7TW0B1fGV-Jve2hU3fKaDyQywsCJp36mpnVaqb5eiTssncHixAwZE0C4yh_XsTd-WoVhsbqlEuDfPTjrtAK94mSzHJTcO3fbtE9L-MoPevQIPM7Yog0i2Xn1oPUCDXVXsV2yJriBiI_r2xbG0nz5Bwn8KAFZ0dNGJ7T9urqKaKMh9guE4jgYLIpRpod_Fd13_GAK0ebgF2CZJdjJT7eEGhzzcg4uFpFdIXL2kNgVN1D6YLMPw3HhVg7_MIfASbJgpcppFhYa4Fk-OpchL5-e_mMyeWogvaJA2wSpyY1f5zJlBnFuIyk_OdV0TwQ3b_TjutehsiibT9WRpOK8h8	eyJhbGciOiJSUzI1NiIsImtpZCI6IldmbUJ1c2Q0RndUVWRmMjc2NHowUWlEYlJ3TnRBcU5pNVlXS1U1c2k0eEEiLCJ0eXAiOiJKV1QifQ.eyJnbnQiOjEsInR0IjoxLCJleHAiOjE2NDQyNTg1MjcsImlhdCI6MTY0MTYzMDUyN30.iVtIGQ6VTgRI8L3xFD_YNvVBGZ6kdFb3ERdyOCIHC_CHhOEpZxVGawMGnNNooqbNdmOqJQ0RLJyiAirEKdxSVrtWvqub6uVMjjpeBylE1sAFymCGNJQf77dKvgPHW3QY5FvOSoOoNcRU2g99Bx8sbZhiI12GnNOB-abazrzICpOUikiTdb2ri3w_TNF2Ibrn-itSa1yuhmTrVpqXt_CT4MEfteiDmgjyqonmk-J_BqbcriF3DKAvrXNK1VKVU7xODcFSIRizlgA2kDmnpMT3Oo-Z1I37TFIGAuDOTgcceOPa7rXg_Mfd_jhL7bSH1BI4RsK0rgde3NaCQlU2n7yVOYGbJCSsSWwSAi-gCjjuTTPnQWe3ep3IWrB73_7tKG2_x7YxZ1nQCSFKouA5rZH4g6yoV8wdJh8_bX2Z64-MJBUl8E7JGM2urA5GY1abo0GZ6ZuQi2JS5WnG1iTL9pFlmOoTpN1DKtNE2PUE90GJwi0qGeACif9uJBXQPDAgKk7fbUxKYQobc6ko2CJ1isoRjbi8-GsJ9lhw7tXno5zfAvN3eps9SYgmIRNh0t_vx-LMBezSTSEcTJpv-7Ap6F10GD3E9KmGcYrOMvdtaYgkWFXO6rh49uElUVid-C1tNVpKjnj7ewUosQo9MHSn-d5l1df0rJSueXcaUMSqRSrEzqQ	1641634127	test@test.test	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	t	OBW2OF5QH3NMCYJ44VU5B5YEQ5LHZLTFW2FDSAJ4R4JVZ4HWSNVQ====	\N	2	1
2	user2	eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMiIsImlhdCI6MTY0MTYzMDUyNywiZXhwIjoxNjQxNjM0MTI3fQ.example_token_user2	eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMiIsImlhdCI6MTY0MTYzMDUyNywiZXhwIjoxNjQ0MjU4NTI3fQ.example_secret_user2	1641634127	user2@test.test	http://10.40.8.5:3000/avatars/default2	f	HASH2EXAMPLEHASH2EXAMPLEHASH2EXAMPLEHASH2EXAMPLE====	\N	2	1
\.


--
-- Data for Name: workflows; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.workflows (id, pipeline_id, pid, name, state, error, started, finished, agent_id, platform, environ, axis_id) FROM stdin;
1	1	1	drone	failure	Error response from daemon: manifest for woodpeckerci/plugin-git:test not found: manifest unknown: manifest unknown	1641630525	1641630527	0		{}	\N
\.


--
-- Name: agents_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.agents_id_seq', 3, true);


--
-- Name: builds_build_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.builds_build_id_seq', 1, true);


--
-- Name: config_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.config_config_id_seq', 1, true);


--
-- Name: crons_i_d_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.crons_i_d_seq', 1, false);


--
-- Name: forge_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.forge_id_seq', 1, true);


--
-- Name: log_entries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.log_entries_id_seq', 1, false);


--
-- Name: orgs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.orgs_id_seq', 2, true);


--
-- Name: procs_proc_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.procs_proc_id_seq', 3, true);


--
-- Name: redirections_redirection_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.redirections_redirection_id_seq', 1, false);


--
-- Name: registry_registry_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.registry_registry_id_seq', 1, false);


--
-- Name: repos_repo_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.repos_repo_id_seq', 122, true);


--
-- Name: secrets_secret_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.secrets_secret_id_seq', 4, true);


--
-- Name: users_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_user_id_seq', 2, true);


--
-- Name: workflows_workflow_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.workflows_workflow_id_seq', 1, true);


--
-- Name: agents agents_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_pkey PRIMARY KEY (id);


--
-- Name: pipelines builds_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipelines
    ADD CONSTRAINT builds_pkey PRIMARY KEY (id);


--
-- Name: configs config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.configs
    ADD CONSTRAINT config_pkey PRIMARY KEY (id);


--
-- Name: crons crons_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.crons
    ADD CONSTRAINT crons_pkey PRIMARY KEY (id);


--
-- Name: forges forge_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forges
    ADD CONSTRAINT forge_pkey PRIMARY KEY (id);


--
-- Name: log_entries log_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.log_entries
    ADD CONSTRAINT log_entries_pkey PRIMARY KEY (id);


--
-- Name: orgs orgs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orgs
    ADD CONSTRAINT orgs_pkey PRIMARY KEY (id);


--
-- Name: perms perms_perm_user_id_perm_repo_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.perms
    ADD CONSTRAINT perms_perm_user_id_perm_repo_id_key UNIQUE (user_id, repo_id);


--
-- Name: steps procs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.steps
    ADD CONSTRAINT procs_pkey PRIMARY KEY (id);


--
-- Name: redirections redirections_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.redirections
    ADD CONSTRAINT redirections_pkey PRIMARY KEY (id);


--
-- Name: registries registry_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.registries
    ADD CONSTRAINT registry_pkey PRIMARY KEY (id);


--
-- Name: repos repos_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repos
    ADD CONSTRAINT repos_pkey PRIMARY KEY (id);


--
-- Name: secrets secrets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.secrets
    ADD CONSTRAINT secrets_pkey PRIMARY KEY (id);


--
-- Name: server_configs server_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.server_configs
    ADD CONSTRAINT server_config_pkey PRIMARY KEY (key);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: workflows workflows_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT workflows_pkey PRIMARY KEY (id);


--
-- Name: IDX_agents_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_agents_org_id" ON public.agents USING btree (org_id);


--
-- Name: IDX_crons_creator_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_crons_creator_id" ON public.crons USING btree (creator_id);


--
-- Name: IDX_crons_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_crons_name" ON public.crons USING btree (name);


--
-- Name: IDX_crons_repo_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_crons_repo_id" ON public.crons USING btree (repo_id);


--
-- Name: IDX_log_entries_step_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_log_entries_step_id" ON public.log_entries USING btree (step_id);


--
-- Name: IDX_perms_perm_repo_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_perms_perm_repo_id" ON public.perms USING btree (repo_id);


--
-- Name: IDX_perms_perm_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_perms_perm_user_id" ON public.perms USING btree (user_id);


--
-- Name: IDX_pipelines_pipeline_author; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_pipelines_pipeline_author" ON public.pipelines USING btree (author);


--
-- Name: IDX_pipelines_pipeline_repo_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_pipelines_pipeline_repo_id" ON public.pipelines USING btree (repo_id);


--
-- Name: IDX_pipelines_pipeline_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_pipelines_pipeline_status" ON public.pipelines USING btree (status);


--
-- Name: IDX_registries_address; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_registries_address" ON public.registries USING btree (address);


--
-- Name: IDX_registries_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_registries_org_id" ON public.registries USING btree (org_id);


--
-- Name: IDX_registries_repo_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_registries_repo_id" ON public.registries USING btree (repo_id);


--
-- Name: IDX_repos_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_repos_org_id" ON public.repos USING btree (org_id);


--
-- Name: IDX_repos_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_repos_user_id" ON public.repos USING btree (user_id);


--
-- Name: IDX_secrets_secret_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_secrets_secret_name" ON public.secrets USING btree (name);


--
-- Name: IDX_secrets_secret_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_secrets_secret_org_id" ON public.secrets USING btree (org_id);


--
-- Name: IDX_secrets_secret_repo_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_secrets_secret_repo_id" ON public.secrets USING btree (repo_id);


--
-- Name: IDX_steps_pipeline_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_steps_pipeline_id" ON public.steps USING btree (pipeline_id);


--
-- Name: IDX_steps_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_steps_uuid" ON public.steps USING btree (uuid);


--
-- Name: IDX_workflows_pipeline_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_workflows_pipeline_id" ON public.workflows USING btree (pipeline_id);


--
-- Name: UQE_config_s; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_config_s" ON public.configs USING btree (repo_id, hash, name);


--
-- Name: UQE_crons_s; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_crons_s" ON public.crons USING btree (name, repo_id);


--
-- Name: UQE_orgs_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_orgs_name" ON public.orgs USING btree (name);


--
-- Name: UQE_pipeline_config_s; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_pipeline_config_s" ON public.pipeline_configs USING btree (config_id, pipeline_id);


--
-- Name: UQE_pipelines_s; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_pipelines_s" ON public.pipelines USING btree (repo_id, number);


--
-- Name: UQE_redirections_repo_full_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_redirections_repo_full_name" ON public.redirections USING btree (repo_full_name);


--
-- Name: UQE_registries_s; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_registries_s" ON public.registries USING btree (org_id, repo_id, address);


--
-- Name: UQE_repos_full_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_repos_full_name" ON public.repos USING btree (full_name);


--
-- Name: UQE_repos_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_repos_name" ON public.repos USING btree (owner, name);


--
-- Name: UQE_secrets_s; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_secrets_s" ON public.secrets USING btree (org_id, repo_id, name);


--
-- Name: UQE_steps_s; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_steps_s" ON public.steps USING btree (pipeline_id, pid);


--
-- Name: UQE_tasks_task_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_tasks_task_id" ON public.tasks USING btree (id);


--
-- Name: UQE_users_hash; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_users_hash" ON public.users USING btree (hash);


--
-- Name: UQE_users_login; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_users_login" ON public.users USING btree (login);


--
-- Name: UQE_workflows_s; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_workflows_s" ON public.workflows USING btree (pipeline_id, pid);


--
-- PostgreSQL database dump complete
--

\unrestrict 18NccrfpihYnwqBivs51uHtRdmaM3893BGxRvmp6T7W9a7eNFNYZo0q0ioWALlU

