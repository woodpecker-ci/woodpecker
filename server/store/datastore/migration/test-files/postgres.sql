--
-- PostgreSQL database dump
--

\restrict jfY4LTom39twz6Gcmw9Je24Z5WfG13hXALefGXbMfzVdfoHA6q0IcaOchrQfKXA

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
    no_schedule boolean
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
    pipeline_id integer NOT NULL,
    pipeline_repo_id integer,
    pipeline_number integer,
    pipeline_event character varying(500),
    pipeline_status character varying(500),
    pipeline_enqueued integer,
    pipeline_created integer,
    pipeline_started integer,
    pipeline_finished integer,
    pipeline_commit character varying(500),
    pipeline_branch character varying(500),
    pipeline_ref character varying(500),
    pipeline_refspec character varying(1000),
    pipeline_clone_url character varying(500),
    pipeline_title character varying(1000),
    pipeline_message text,
    pipeline_timestamp integer,
    pipeline_author character varying(500),
    pipeline_avatar character varying(1000),
    pipeline_email character varying(500),
    pipeline_forge_url character varying(1000),
    pipeline_deploy character varying(500),
    pipeline_parent integer,
    pipeline_reviewer character varying(250),
    pipeline_reviewed integer,
    pipeline_sender character varying(250),
    pipeline_config_id integer,
    changed_files text,
    updated bigint DEFAULT 0 NOT NULL,
    additional_variables json,
    pr_labels json,
    pipeline_errors json
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

ALTER SEQUENCE public.builds_build_id_seq OWNED BY public.pipelines.pipeline_id;


--
-- Name: config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.config (
    config_id integer NOT NULL,
    config_repo_id integer,
    config_hash character varying(250),
    config_data bytea,
    config_name text
);


ALTER TABLE public.config OWNER TO postgres;

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

ALTER SEQUENCE public.config_config_id_seq OWNED BY public.config.config_id;


--
-- Name: crons; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.crons (
    i_d bigint NOT NULL,
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

ALTER SEQUENCE public.crons_i_d_seq OWNED BY public.crons.i_d;


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
-- Name: migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.migrations (
    name character varying(255)
);


ALTER TABLE public.migrations OWNER TO postgres;

--
-- Name: orgs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.orgs (
    id bigint NOT NULL,
    name character varying(255),
    is_user boolean,
    private boolean
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
    perm_user_id integer NOT NULL,
    perm_repo_id integer NOT NULL,
    perm_pull boolean,
    perm_push boolean,
    perm_admin boolean,
    perm_synced integer,
    created bigint,
    updated bigint
);


ALTER TABLE public.perms OWNER TO postgres;

--
-- Name: pipeline_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pipeline_config (
    config_id bigint NOT NULL,
    pipeline_id bigint NOT NULL
);


ALTER TABLE public.pipeline_config OWNER TO postgres;

--
-- Name: steps; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.steps (
    step_id integer NOT NULL,
    step_pipeline_id integer,
    step_pid integer,
    step_ppid integer,
    step_name character varying(250),
    step_state character varying(250),
    step_error text,
    step_exit_code integer,
    step_started integer,
    step_stopped integer,
    step_machine character varying(250),
    step_uuid character varying(255),
    step_failure character varying(255),
    step_type character varying(255)
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

ALTER SEQUENCE public.procs_proc_id_seq OWNED BY public.steps.step_id;


--
-- Name: redirections; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.redirections (
    redirection_id bigint NOT NULL,
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

ALTER SEQUENCE public.redirections_redirection_id_seq OWNED BY public.redirections.redirection_id;


--
-- Name: registry; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.registry (
    registry_id integer NOT NULL,
    registry_repo_id integer,
    registry_addr character varying(250),
    registry_email character varying(500),
    registry_username character varying(2000),
    registry_password text,
    registry_token text
);


ALTER TABLE public.registry OWNER TO postgres;

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

ALTER SEQUENCE public.registry_registry_id_seq OWNED BY public.registry.registry_id;


--
-- Name: repos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.repos (
    repo_id integer NOT NULL,
    repo_user_id integer,
    repo_owner character varying(250),
    repo_name character varying(250),
    repo_full_name character varying(250),
    repo_avatar character varying(500),
    repo_forge_url character varying(1000),
    repo_clone character varying(1000),
    repo_branch character varying(500),
    repo_timeout integer,
    repo_private boolean,
    repo_trusted boolean,
    repo_allow_pr boolean,
    repo_allow_push boolean,
    repo_hash character varying(500),
    repo_scm character varying(50),
    repo_config_path character varying(500),
    repo_gated boolean,
    repo_visibility character varying(50),
    repo_active boolean,
    forge_remote_id character varying(255),
    repo_org_id bigint,
    cancel_previous_pipeline_events json,
    netrc_only_trusted boolean DEFAULT true NOT NULL,
    repo_clone_ssh character varying(1000)
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

ALTER SEQUENCE public.repos_repo_id_seq OWNED BY public.repos.repo_id;


--
-- Name: secrets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.secrets (
    secret_id integer NOT NULL,
    secret_repo_id integer DEFAULT 0 NOT NULL,
    secret_name character varying(250) NOT NULL,
    secret_value bytea,
    secret_images character varying(2000),
    secret_events character varying(2000),
    secret_org_id bigint DEFAULT 0 NOT NULL
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

ALTER SEQUENCE public.secrets_secret_id_seq OWNED BY public.secrets.secret_id;


--
-- Name: server_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.server_config (
    key character varying(255) NOT NULL,
    value character varying(255)
);


ALTER TABLE public.server_config OWNER TO postgres;

--
-- Name: tasks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tasks (
    task_id character varying(250) NOT NULL,
    task_data bytea,
    task_labels bytea,
    task_dependencies bytea,
    task_run_on bytea,
    task_dep_status json,
    agent_id bigint
);


ALTER TABLE public.tasks OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    user_id integer NOT NULL,
    user_login character varying(250),
    user_token text,
    user_secret text,
    user_expiry integer,
    user_email character varying(500),
    user_avatar character varying(500),
    user_admin boolean,
    user_hash character varying(500),
    forge_remote_id character varying(255),
    user_org_id bigint
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

ALTER SEQUENCE public.users_user_id_seq OWNED BY public.users.user_id;


--
-- Name: workflows; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.workflows (
    workflow_id bigint NOT NULL,
    workflow_pipeline_id bigint,
    workflow_pid integer,
    workflow_name character varying(255),
    workflow_state character varying(255),
    workflow_error text,
    workflow_started bigint,
    workflow_stopped bigint,
    workflow_agent_id bigint,
    workflow_platform character varying(255),
    workflow_environ json,
    workflow_axis_id integer
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

ALTER SEQUENCE public.workflows_workflow_id_seq OWNED BY public.workflows.workflow_id;


--
-- Name: agents id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.agents ALTER COLUMN id SET DEFAULT nextval('public.agents_id_seq'::regclass);


--
-- Name: config config_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.config ALTER COLUMN config_id SET DEFAULT nextval('public.config_config_id_seq'::regclass);


--
-- Name: crons i_d; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.crons ALTER COLUMN i_d SET DEFAULT nextval('public.crons_i_d_seq'::regclass);


--
-- Name: log_entries id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.log_entries ALTER COLUMN id SET DEFAULT nextval('public.log_entries_id_seq'::regclass);


--
-- Name: orgs id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orgs ALTER COLUMN id SET DEFAULT nextval('public.orgs_id_seq'::regclass);


--
-- Name: pipelines pipeline_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipelines ALTER COLUMN pipeline_id SET DEFAULT nextval('public.builds_build_id_seq'::regclass);


--
-- Name: redirections redirection_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.redirections ALTER COLUMN redirection_id SET DEFAULT nextval('public.redirections_redirection_id_seq'::regclass);


--
-- Name: registry registry_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.registry ALTER COLUMN registry_id SET DEFAULT nextval('public.registry_registry_id_seq'::regclass);


--
-- Name: repos repo_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repos ALTER COLUMN repo_id SET DEFAULT nextval('public.repos_repo_id_seq'::regclass);


--
-- Name: secrets secret_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.secrets ALTER COLUMN secret_id SET DEFAULT nextval('public.secrets_secret_id_seq'::regclass);


--
-- Name: steps step_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.steps ALTER COLUMN step_id SET DEFAULT nextval('public.procs_proc_id_seq'::regclass);


--
-- Name: users user_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN user_id SET DEFAULT nextval('public.users_user_id_seq'::regclass);


--
-- Name: workflows workflow_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.workflows ALTER COLUMN workflow_id SET DEFAULT nextval('public.workflows_workflow_id_seq'::regclass);


--
-- Data for Name: agents; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.agents (id, created, updated, name, owner_id, token, last_contact, platform, backend, capacity, version, no_schedule) FROM stdin;
\.


--
-- Data for Name: config; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.config (config_id, config_repo_id, config_hash, config_data, config_name) FROM stdin;
1	105	ec8ca9529d6081e631aec26175b26ac91699395b96b9c5fc1f3af6d3aef5d3a8	\\x636c6f6e653a0a20206769743a0a20202020696d6167653a20776f6f647065636b657263692f706c7567696e2d6769743a746573740a0a706970656c696e653a0a20205072696e743a0a20202020696d6167653a207072696e742f656e760a20202020736563726574733a205b204141414141414141414141414141414141414141414141414141205d	drone
\.


--
-- Data for Name: crons; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.crons (i_d, name, repo_id, creator_id, next_exec, schedule, created, branch) FROM stdin;
\.


--
-- Data for Name: log_entries; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.log_entries (id, step_id, "time", line, data, created, type) FROM stdin;
\.


--
-- Data for Name: migrations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.migrations (name) FROM stdin;
create-table-users
create-table-repos
create-table-builds
create-index-builds-repo
create-index-builds-author
create-table-procs
create-index-procs-build
create-table-logs
create-table-files
create-index-files-builds
create-index-files-procs
create-table-secrets
create-index-secrets-repo
create-table-registry
create-index-registry-repo
create-table-config
create-table-tasks
create-table-agents
create-table-senders
create-index-sender-repos
alter-table-add-repo-visibility
update-table-set-repo-visibility
alter-table-add-repo-seq
update-table-set-repo-seq
update-table-set-repo-seq-default
alter-table-add-repo-active
update-table-set-repo-active
alter-table-add-user-synced
update-table-set-user-synced
create-table-perms
create-index-perms-repo
create-index-perms-user
alter-table-add-file-pid
alter-table-add-file-meta-passed
alter-table-add-file-meta-failed
alter-table-add-file-meta-skipped
alter-table-update-file-meta
create-table-build-config
alter-table-add-config-name
update-table-set-config-name
populate-build-config
alter-table-add-task-dependencies
alter-table-add-task-run-on
alter-table-add-repo-fallback
update-table-set-repo-fallback
update-table-set-repo-fallback-again
add-builds-changed_files-column
update-builds-set-changed_files
update-table-set-users-token-and-secret-length
xorm
alter-table-drop-repo-fallback
drop-allow-push-tags-deploys-columns
alter-table-drop-counter
drop-senders
alter-table-logs-update-type-of-data
alter-table-add-secrets-user-id
recreate-agents-table
lowercase-secret-names
rename-builds-to-pipeline
rename-columns-builds-to-pipeline
rename-procs-to-steps
rename-remote-to-forge
rename-forge-id-to-forge-remote-id
remove-active-from-users
remove-inactive-repos
drop-files
init-log_entries
migrate-logs-to-log_entries
parent-steps-to-workflows
add-orgs
add-org-id
alter-table-tasks-update-type-of-task-data
alter-table-config-update-type-of-config-data
remove-plugin-only-option-from-secrets-table
drop-old-col
convert-to-new-pipeline-error-format
rename-link-to-url
\.


--
-- Data for Name: orgs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.orgs (id, name, is_user, private) FROM stdin;
1	2	f	f
2	test	t	f
\.


--
-- Data for Name: perms; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.perms (perm_user_id, perm_repo_id, perm_pull, perm_push, perm_admin, perm_synced, created, updated) FROM stdin;
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
-- Data for Name: pipeline_config; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pipeline_config (config_id, pipeline_id) FROM stdin;
1	1
\.


--
-- Data for Name: pipelines; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pipelines (pipeline_id, pipeline_repo_id, pipeline_number, pipeline_event, pipeline_status, pipeline_enqueued, pipeline_created, pipeline_started, pipeline_finished, pipeline_commit, pipeline_branch, pipeline_ref, pipeline_refspec, pipeline_clone_url, pipeline_title, pipeline_message, pipeline_timestamp, pipeline_author, pipeline_avatar, pipeline_email, pipeline_forge_url, pipeline_deploy, pipeline_parent, pipeline_reviewer, pipeline_reviewed, pipeline_sender, pipeline_config_id, changed_files, updated, additional_variables, pr_labels, pipeline_errors) FROM stdin;
1	105	1	push	failure	1641630525	1641630525	1641630525	1641630527	24bf205107cea48b92bc6444e18e40d21733a594	master	refs/heads/master				„.drone.yml“ hinzufügen\\n	1641630525	test	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	test@test.test	http://10.40.8.5:3000/2/settings/compare/3fee083df05667d525878b5fcbd4eaf2a121c559...24bf205107cea48b92bc6444e18e40d21733a594		0		0	test	0	[".drone.yml"]\\n	0	\N	\N	\N
\.


--
-- Data for Name: redirections; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.redirections (redirection_id, repo_id, repo_full_name) FROM stdin;
\.


--
-- Data for Name: registry; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.registry (registry_id, registry_repo_id, registry_addr, registry_email, registry_username, registry_password, registry_token) FROM stdin;
\.


--
-- Data for Name: repos; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.repos (repo_id, repo_user_id, repo_owner, repo_name, repo_full_name, repo_avatar, repo_forge_url, repo_clone, repo_branch, repo_timeout, repo_private, repo_trusted, repo_allow_pr, repo_allow_push, repo_hash, repo_scm, repo_config_path, repo_gated, repo_visibility, repo_active, forge_remote_id, repo_org_id, cancel_previous_pipeline_events, netrc_only_trusted, repo_clone_ssh) FROM stdin;
115	1	2	testCIservices	2/testCIservices	http://10.40.8.5:3000/avatars/c81e728d9d4c2f636f067f89cc14862c	http://10.40.8.5:3000/2/testCIservices	http://10.40.8.5:3000/2/testCIservices.git	master	60	f	f	t	t	FOUXTSNL2GXK7JP2SQQJVWVAS6J4E4SGIQYPAHEJBIFPVR46LLDA====	git	.drone.yml	f	public	t	\N	1	\N	t	\N
105	1	2	settings	2/settings	http://10.40.8.5:3000/avatars/c81e728d9d4c2f636f067f89cc14862c	http://10.40.8.5:3000/2/settings	http://10.40.8.5:3000/2/settings.git	master	60	f	f	t	t	3OQA7X5CNGPTILDYLQSJFDML6U2W7UUFBPPP2G2LRBG3WETAYZLA====	git	.drone.yml	f	public	t	\N	1	\N	t	\N
\.


--
-- Data for Name: secrets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.secrets (secret_id, secret_repo_id, secret_name, secret_value, secret_images, secret_events, secret_org_id) FROM stdin;
1	105	wow	\\x74657374	null\\n	["push","tag","deployment","pull_request"]\\n	0
2	105	n	\\x6e	null\\n	["deployment"]\\n	0
3	105	abc	\\x656466	null\\n	["push"]\\n	0
4	105	quak	\\x66647361	null\\n	["pull-request"]\\n	0
\.


--
-- Data for Name: server_config; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.server_config (key, value) FROM stdin;
signature-private-key	1fe3b71c87d7f89fa878306028cf08d66020ef6cafc2af90d05c40ebd03eee3c93189d2a3c46fe5292afc33e9237615ed595ee3d588dce431d5f6848b6a9bf77
\.


--
-- Data for Name: steps; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.steps (step_id, step_pipeline_id, step_pid, step_ppid, step_name, step_state, step_error, step_exit_code, step_started, step_stopped, step_machine, step_uuid, step_failure, step_type) FROM stdin;
2	1	2	1	git	success		0	1641630525	1641630527	someHostname	\N	\N	\N
3	1	3	1	Print	skipped		0	0	0		\N	\N	\N
\.


--
-- Data for Name: tasks; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tasks (task_id, task_data, task_labels, task_dependencies, task_run_on, task_dep_status, agent_id) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (user_id, user_login, user_token, user_secret, user_expiry, user_email, user_avatar, user_admin, user_hash, forge_remote_id, user_org_id) FROM stdin;
1	test	eyJhbGciOiJSUzI1NiIsImtpZCI6IldmbUJ1c2Q0RndUVWRmMjc2NHowUWlEYlJ3TnRBcU5pNVlXS1U1c2k0eEEiLCJ0eXAiOiJKV1QifQ.eyJnbnQiOjEsInR0IjowLCJleHAiOjE2NDE2MzQxMjcsImlhdCI6MTY0MTYzMDUyN30.Fu0wUP-08NpPjq737y6HOeyKN_-_SE4iOZr5yrH7S8Jrz8nIuNKfU7AvlypeMSJ7wo8e3cSTadbSH1polZuFv-Nb1AqWDDXeuXudm61BkF96sTslbSHd0nF7cOy6hqCfIAfQLQpqZTJZ4E26oOSSJxPfOOntOWhlEejRl5F-flXAoYAQLegHxdn9IfYJeM1eanZqF4k6dT9hthFp9v4fmUjODPPfHip_iS7ckPonP1E4-8KeNkU3O-lIS1fgrsbCDA8531FXIGB0U7cSur7H0picKGL6WSzAErPGntlNlQWYB5JedDtLN9Ionxy1Y9LKQON76XYL4gM1Ji98RCEXggVqd7TW0B1fGV-Jve2hU3fKaDyQywsCJp36mpnVaqb5eiTssncHixAwZE0C4yh_XsTd-WoVhsbqlEuDfPTjrtAK94mSzHJTcO3fbtE9L-MoPevQIPM7Yog0i2Xn1oPUCDXVXsV2yJriBiI_r2xbG0nz5Bwn8KAFZ0dNGJ7T9urqKaKMh9guE4jgYLIpRpod_Fd13_GAK0ebgF2CZJdjJT7eEGhzzcg4uFpFdIXL2kNgVN1D6YLMPw3HhVg7_MIfASbJgpcppFhYa4Fk-OpchL5-e_mMyeWogvaJA2wSpyY1f5zJlBnFuIyk_OdV0TwQ3b_TjutehsiibT9WRpOK8h8	eyJhbGciOiJSUzI1NiIsImtpZCI6IldmbUJ1c2Q0RndUVWRmMjc2NHowUWlEYlJ3TnRBcU5pNVlXS1U1c2k0eEEiLCJ0eXAiOiJKV1QifQ.eyJnbnQiOjEsInR0IjoxLCJleHAiOjE2NDQyNTg1MjcsImlhdCI6MTY0MTYzMDUyN30.iVtIGQ6VTgRI8L3xFD_YNvVBGZ6kdFb3ERdyOCIHC_CHhOEpZxVGawMGnNNooqbNdmOqJQ0RLJyiAirEKdxSVrtWvqub6uVMjjpeBylE1sAFymCGNJQf77dKvgPHW3QY5FvOSoOoNcRU2g99Bx8sbZhiI12GnNOB-abazrzICpOUikiTdb2ri3w_TNF2Ibrn-itSa1yuhmTrVpqXt_CT4MEfteiDmgjyqonmk-J_BqbcriF3DKAvrXNK1VKVU7xODcFSIRizlgA2kDmnpMT3Oo-Z1I37TFIGAuDOTgcceOPa7rXg_Mfd_jhL7bSH1BI4RsK0rgde3NaCQlU2n7yVOYGbJCSsSWwSAi-gCjjuTTPnQWe3ep3IWrB73_7tKG2_x7YxZ1nQCSFKouA5rZH4g6yoV8wdJh8_bX2Z64-MJBUl8E7JGM2urA5GY1abo0GZ6ZuQi2JS5WnG1iTL9pFlmOoTpN1DKtNE2PUE90GJwi0qGeACif9uJBXQPDAgKk7fbUxKYQobc6ko2CJ1isoRjbi8-GsJ9lhw7tXno5zfAvN3eps9SYgmIRNh0t_vx-LMBezSTSEcTJpv-7Ap6F10GD3E9KmGcYrOMvdtaYgkWFXO6rh49uElUVid-C1tNVpKjnj7ewUosQo9MHSn-d5l1df0rJSueXcaUMSqRSrEzqQ	1641634127	test@test.test	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	f	OBW2OF5QH3NMCYJ44VU5B5YEQ5LHZLTFW2FDSAJ4R4JVZ4HWSNVQ====	\N	2
\.


--
-- Data for Name: workflows; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.workflows (workflow_id, workflow_pipeline_id, workflow_pid, workflow_name, workflow_state, workflow_error, workflow_started, workflow_stopped, workflow_agent_id, workflow_platform, workflow_environ, workflow_axis_id) FROM stdin;
1	1	1	drone	failure	Error response from daemon: manifest for woodpeckerci/plugin-git:test not found: manifest unknown: manifest unknown	1641630525	1641630527	0		{}	\N
\.


--
-- Name: agents_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.agents_id_seq', 1, false);


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

SELECT pg_catalog.setval('public.users_user_id_seq', 1, true);


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
    ADD CONSTRAINT builds_pkey PRIMARY KEY (pipeline_id);


--
-- Name: config config_config_hash_config_repo_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.config
    ADD CONSTRAINT config_config_hash_config_repo_id_key UNIQUE (config_hash, config_repo_id);


--
-- Name: config config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.config
    ADD CONSTRAINT config_pkey PRIMARY KEY (config_id);


--
-- Name: crons crons_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.crons
    ADD CONSTRAINT crons_pkey PRIMARY KEY (i_d);


--
-- Name: log_entries log_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.log_entries
    ADD CONSTRAINT log_entries_pkey PRIMARY KEY (id);


--
-- Name: migrations migrations_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT migrations_name_key UNIQUE (name);


--
-- Name: orgs orgs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orgs
    ADD CONSTRAINT orgs_pkey PRIMARY KEY (id);


--
-- Name: perms perms_perm_user_id_perm_repo_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.perms
    ADD CONSTRAINT perms_perm_user_id_perm_repo_id_key UNIQUE (perm_user_id, perm_repo_id);


--
-- Name: steps procs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.steps
    ADD CONSTRAINT procs_pkey PRIMARY KEY (step_id);


--
-- Name: steps procs_proc_build_id_proc_pid_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.steps
    ADD CONSTRAINT procs_proc_build_id_proc_pid_key UNIQUE (step_pipeline_id, step_pid);


--
-- Name: redirections redirections_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.redirections
    ADD CONSTRAINT redirections_pkey PRIMARY KEY (redirection_id);


--
-- Name: registry registry_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.registry
    ADD CONSTRAINT registry_pkey PRIMARY KEY (registry_id);


--
-- Name: registry registry_registry_addr_registry_repo_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.registry
    ADD CONSTRAINT registry_registry_addr_registry_repo_id_key UNIQUE (registry_addr, registry_repo_id);


--
-- Name: repos repos_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repos
    ADD CONSTRAINT repos_pkey PRIMARY KEY (repo_id);


--
-- Name: secrets secrets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.secrets
    ADD CONSTRAINT secrets_pkey PRIMARY KEY (secret_id);


--
-- Name: server_config server_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.server_config
    ADD CONSTRAINT server_config_pkey PRIMARY KEY (key);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (task_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- Name: users users_user_login_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_user_login_key UNIQUE (user_login);


--
-- Name: workflows workflows_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT workflows_pkey PRIMARY KEY (workflow_id);


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

CREATE INDEX "IDX_perms_perm_repo_id" ON public.perms USING btree (perm_repo_id);


--
-- Name: IDX_perms_perm_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_perms_perm_user_id" ON public.perms USING btree (perm_user_id);


--
-- Name: IDX_pipelines_pipeline_author; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_pipelines_pipeline_author" ON public.pipelines USING btree (pipeline_author);


--
-- Name: IDX_pipelines_pipeline_repo_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_pipelines_pipeline_repo_id" ON public.pipelines USING btree (pipeline_repo_id);


--
-- Name: IDX_pipelines_pipeline_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_pipelines_pipeline_status" ON public.pipelines USING btree (pipeline_status);


--
-- Name: IDX_registry_registry_addr; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_registry_registry_addr" ON public.registry USING btree (registry_addr);


--
-- Name: IDX_registry_registry_repo_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_registry_registry_repo_id" ON public.registry USING btree (registry_repo_id);


--
-- Name: IDX_secrets_secret_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_secrets_secret_name" ON public.secrets USING btree (secret_name);


--
-- Name: IDX_secrets_secret_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_secrets_secret_org_id" ON public.secrets USING btree (secret_org_id);


--
-- Name: IDX_secrets_secret_repo_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_secrets_secret_repo_id" ON public.secrets USING btree (secret_repo_id);


--
-- Name: IDX_steps_step_pipeline_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_steps_step_pipeline_id" ON public.steps USING btree (step_pipeline_id);


--
-- Name: IDX_steps_step_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_steps_step_uuid" ON public.steps USING btree (step_uuid);


--
-- Name: IDX_workflows_workflow_pipeline_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_workflows_workflow_pipeline_id" ON public.workflows USING btree (workflow_pipeline_id);


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

CREATE UNIQUE INDEX "UQE_pipeline_config_s" ON public.pipeline_config USING btree (config_id, pipeline_id);


--
-- Name: UQE_pipelines_s; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_pipelines_s" ON public.pipelines USING btree (pipeline_repo_id, pipeline_number);


--
-- Name: UQE_redirections_repo_full_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_redirections_repo_full_name" ON public.redirections USING btree (repo_full_name);


--
-- Name: UQE_repos_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_repos_name" ON public.repos USING btree (repo_owner, repo_name);


--
-- Name: UQE_repos_repo_full_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_repos_repo_full_name" ON public.repos USING btree (repo_full_name);


--
-- Name: UQE_secrets_s; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_secrets_s" ON public.secrets USING btree (secret_org_id, secret_repo_id, secret_name);


--
-- Name: UQE_tasks_task_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_tasks_task_id" ON public.tasks USING btree (task_id);


--
-- Name: UQE_users_user_hash; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_users_user_hash" ON public.users USING btree (user_hash);


--
-- Name: UQE_workflows_s; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_workflows_s" ON public.workflows USING btree (workflow_pipeline_id, workflow_pid);


--
-- PostgreSQL database dump complete
--

\unrestrict jfY4LTom39twz6Gcmw9Je24Z5WfG13hXALefGXbMfzVdfoHA6q0IcaOchrQfKXA

