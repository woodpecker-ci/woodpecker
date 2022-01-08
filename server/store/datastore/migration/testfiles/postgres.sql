--
-- PostgreSQL database dump
--

-- Dumped from database version 13.4
-- Dumped by pg_dump version 13.4

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: agents; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.agents (
    agent_id integer NOT NULL,
    agent_addr character varying(250),
    agent_platform character varying(500),
    agent_capacity integer,
    agent_created integer,
    agent_updated integer
);


ALTER TABLE public.agents OWNER TO root;

--
-- Name: agents_agent_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.agents_agent_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.agents_agent_id_seq OWNER TO root;

--
-- Name: agents_agent_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.agents_agent_id_seq OWNED BY public.agents.agent_id;


--
-- Name: build_config; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.build_config (
    config_id integer NOT NULL,
    build_id integer NOT NULL
);


ALTER TABLE public.build_config OWNER TO root;

--
-- Name: builds; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.builds (
    build_id integer NOT NULL,
    build_repo_id integer,
    build_number integer,
    build_event character varying(500),
    build_status character varying(500),
    build_enqueued integer,
    build_created integer,
    build_started integer,
    build_finished integer,
    build_commit character varying(500),
    build_branch character varying(500),
    build_ref character varying(500),
    build_refspec character varying(1000),
    build_remote character varying(500),
    build_title character varying(1000),
    build_message character varying(2000),
    build_timestamp integer,
    build_author character varying(500),
    build_avatar character varying(1000),
    build_email character varying(500),
    build_link character varying(1000),
    build_deploy character varying(500),
    build_signed boolean,
    build_verified boolean,
    build_parent integer,
    build_error character varying(500),
    build_reviewer character varying(250),
    build_reviewed integer,
    build_sender character varying(250),
    build_config_id integer,
    changed_files text
);


ALTER TABLE public.builds OWNER TO root;

--
-- Name: builds_build_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.builds_build_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.builds_build_id_seq OWNER TO root;

--
-- Name: builds_build_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.builds_build_id_seq OWNED BY public.builds.build_id;


--
-- Name: config; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.config (
    config_id integer NOT NULL,
    config_repo_id integer,
    config_hash character varying(250),
    config_data bytea,
    config_name text
);


ALTER TABLE public.config OWNER TO root;

--
-- Name: config_config_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.config_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.config_config_id_seq OWNER TO root;

--
-- Name: config_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.config_config_id_seq OWNED BY public.config.config_id;


--
-- Name: files; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.files (
    file_id integer NOT NULL,
    file_build_id integer,
    file_proc_id integer,
    file_name character varying(250),
    file_mime character varying(250),
    file_size integer,
    file_time integer,
    file_data bytea,
    file_pid integer,
    file_meta_passed integer,
    file_meta_failed integer,
    file_meta_skipped integer
);


ALTER TABLE public.files OWNER TO root;

--
-- Name: files_file_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.files_file_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.files_file_id_seq OWNER TO root;

--
-- Name: files_file_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.files_file_id_seq OWNED BY public.files.file_id;


--
-- Name: logs; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.logs (
    log_id integer NOT NULL,
    log_job_id integer,
    log_data bytea
);


ALTER TABLE public.logs OWNER TO root;

--
-- Name: logs_log_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.logs_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.logs_log_id_seq OWNER TO root;

--
-- Name: logs_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.logs_log_id_seq OWNED BY public.logs.log_id;


--
-- Name: migrations; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.migrations (
    name character varying(255)
);


ALTER TABLE public.migrations OWNER TO root;

--
-- Name: perms; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.perms (
    perm_user_id integer NOT NULL,
    perm_repo_id integer NOT NULL,
    perm_pull boolean,
    perm_push boolean,
    perm_admin boolean,
    perm_synced integer
);


ALTER TABLE public.perms OWNER TO root;

--
-- Name: procs; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.procs (
    proc_id integer NOT NULL,
    proc_build_id integer,
    proc_pid integer,
    proc_ppid integer,
    proc_pgid integer,
    proc_name character varying(250),
    proc_state character varying(250),
    proc_error character varying(500),
    proc_exit_code integer,
    proc_started integer,
    proc_stopped integer,
    proc_machine character varying(250),
    proc_platform character varying(250),
    proc_environ character varying(2000)
);


ALTER TABLE public.procs OWNER TO root;

--
-- Name: procs_proc_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.procs_proc_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.procs_proc_id_seq OWNER TO root;

--
-- Name: procs_proc_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.procs_proc_id_seq OWNED BY public.procs.proc_id;


--
-- Name: registry; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.registry (
    registry_id integer NOT NULL,
    registry_repo_id integer,
    registry_addr character varying(250),
    registry_email character varying(500),
    registry_username character varying(2000),
    registry_password character varying(8000),
    registry_token character varying(2000)
);


ALTER TABLE public.registry OWNER TO root;

--
-- Name: registry_registry_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.registry_registry_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.registry_registry_id_seq OWNER TO root;

--
-- Name: registry_registry_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.registry_registry_id_seq OWNED BY public.registry.registry_id;


--
-- Name: repos; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.repos (
    repo_id integer NOT NULL,
    repo_user_id integer,
    repo_owner character varying(250),
    repo_name character varying(250),
    repo_full_name character varying(250),
    repo_avatar character varying(500),
    repo_link character varying(1000),
    repo_clone character varying(1000),
    repo_branch character varying(500),
    repo_timeout integer,
    repo_private boolean,
    repo_trusted boolean,
    repo_allow_pr boolean,
    repo_allow_push boolean,
    repo_allow_deploys boolean,
    repo_allow_tags boolean,
    repo_hash character varying(500),
    repo_scm character varying(50),
    repo_config_path character varying(500),
    repo_gated boolean,
    repo_visibility character varying(50),
    repo_counter integer,
    repo_active boolean,
    repo_fallback boolean
);


ALTER TABLE public.repos OWNER TO root;

--
-- Name: repos_repo_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.repos_repo_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.repos_repo_id_seq OWNER TO root;

--
-- Name: repos_repo_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.repos_repo_id_seq OWNED BY public.repos.repo_id;


--
-- Name: secrets; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.secrets (
    secret_id integer NOT NULL,
    secret_repo_id integer,
    secret_name character varying(250),
    secret_value bytea,
    secret_images character varying(2000),
    secret_events character varying(2000),
    secret_skip_verify boolean,
    secret_conceal boolean
);


ALTER TABLE public.secrets OWNER TO root;

--
-- Name: secrets_secret_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.secrets_secret_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.secrets_secret_id_seq OWNER TO root;

--
-- Name: secrets_secret_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.secrets_secret_id_seq OWNED BY public.secrets.secret_id;


--
-- Name: senders; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.senders (
    sender_id integer NOT NULL,
    sender_repo_id integer,
    sender_login character varying(250),
    sender_allow boolean,
    sender_block boolean
);


ALTER TABLE public.senders OWNER TO root;

--
-- Name: senders_sender_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.senders_sender_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.senders_sender_id_seq OWNER TO root;

--
-- Name: senders_sender_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.senders_sender_id_seq OWNED BY public.senders.sender_id;


--
-- Name: tasks; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.tasks (
    task_id character varying(250) NOT NULL,
    task_data bytea,
    task_labels bytea,
    task_dependencies bytea,
    task_run_on bytea
);


ALTER TABLE public.tasks OWNER TO root;

--
-- Name: users; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.users (
    user_id integer NOT NULL,
    user_login character varying(250),
    user_token character varying(1000),
    user_secret character varying(1000),
    user_expiry integer,
    user_email character varying(500),
    user_avatar character varying(500),
    user_active boolean,
    user_admin boolean,
    user_hash character varying(500),
    user_synced integer
);


ALTER TABLE public.users OWNER TO root;

--
-- Name: users_user_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.users_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_user_id_seq OWNER TO root;

--
-- Name: users_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.users_user_id_seq OWNED BY public.users.user_id;


--
-- Name: agents agent_id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.agents ALTER COLUMN agent_id SET DEFAULT nextval('public.agents_agent_id_seq'::regclass);


--
-- Name: builds build_id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.builds ALTER COLUMN build_id SET DEFAULT nextval('public.builds_build_id_seq'::regclass);


--
-- Name: config config_id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.config ALTER COLUMN config_id SET DEFAULT nextval('public.config_config_id_seq'::regclass);


--
-- Name: files file_id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.files ALTER COLUMN file_id SET DEFAULT nextval('public.files_file_id_seq'::regclass);


--
-- Name: logs log_id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.logs ALTER COLUMN log_id SET DEFAULT nextval('public.logs_log_id_seq'::regclass);


--
-- Name: procs proc_id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.procs ALTER COLUMN proc_id SET DEFAULT nextval('public.procs_proc_id_seq'::regclass);


--
-- Name: registry registry_id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.registry ALTER COLUMN registry_id SET DEFAULT nextval('public.registry_registry_id_seq'::regclass);


--
-- Name: repos repo_id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.repos ALTER COLUMN repo_id SET DEFAULT nextval('public.repos_repo_id_seq'::regclass);


--
-- Name: secrets secret_id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.secrets ALTER COLUMN secret_id SET DEFAULT nextval('public.secrets_secret_id_seq'::regclass);


--
-- Name: senders sender_id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.senders ALTER COLUMN sender_id SET DEFAULT nextval('public.senders_sender_id_seq'::regclass);


--
-- Name: users user_id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.users ALTER COLUMN user_id SET DEFAULT nextval('public.users_user_id_seq'::regclass);


--
-- Data for Name: agents; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.agents (agent_id, agent_addr, agent_platform, agent_capacity, agent_created, agent_updated) FROM stdin;
\.


--
-- Data for Name: build_config; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.build_config (config_id, build_id) FROM stdin;
1	1
\.


--
-- Data for Name: builds; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.builds (build_id, build_repo_id, build_number, build_event, build_status, build_enqueued, build_created, build_started, build_finished, build_commit, build_branch, build_ref, build_refspec, build_remote, build_title, build_message, build_timestamp, build_author, build_avatar, build_email, build_link, build_deploy, build_signed, build_verified, build_parent, build_error, build_reviewer, build_reviewed, build_sender, build_config_id, changed_files) FROM stdin;
1	105	1	push	failure	1641630525	1641630525	1641630525	1641630527	24bf205107cea48b92bc6444e18e40d21733a594	master	refs/heads/master				„.drone.yml“ hinzufügen\n	1641630525	test	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	test@test.test	http://10.40.8.5:3000/2/settings/compare/3fee083df05667d525878b5fcbd4eaf2a121c559...24bf205107cea48b92bc6444e18e40d21733a594		f	t	0			0	test	0	[".drone.yml"]\n
\.


--
-- Data for Name: config; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.config (config_id, config_repo_id, config_hash, config_data, config_name) FROM stdin;
1	105	ec8ca9529d6081e631aec26175b26ac91699395b96b9c5fc1f3af6d3aef5d3a8	\\x636c6f6e653a0a20206769743a0a20202020696d6167653a20776f6f647065636b657263692f706c7567696e2d6769743a746573740a0a706970656c696e653a0a20205072696e743a0a20202020696d6167653a207072696e742f656e760a20202020736563726574733a205b204141414141414141414141414141414141414141414141414141205d	drone
\.


--
-- Data for Name: files; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.files (file_id, file_build_id, file_proc_id, file_name, file_mime, file_size, file_time, file_data, file_pid, file_meta_passed, file_meta_failed, file_meta_skipped) FROM stdin;
\.


--
-- Data for Name: logs; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.logs (log_id, log_job_id, log_data) FROM stdin;
\.


--
-- Data for Name: migrations; Type: TABLE DATA; Schema: public; Owner: root
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
\.


--
-- Data for Name: perms; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.perms (perm_user_id, perm_repo_id, perm_pull, perm_push, perm_admin, perm_synced) FROM stdin;
1	1	t	t	t	1641626844
1	2	t	t	t	1641626844
1	3	t	t	t	1641626844
1	4	t	t	t	1641626844
1	5	t	t	t	1641626844
1	6	t	t	t	1641626844
1	7	t	t	t	1641626844
1	8	t	t	t	1641626844
1	9	t	t	t	1641626844
1	10	t	t	t	1641626844
1	11	t	t	t	1641626844
1	12	t	t	t	1641626844
1	13	t	t	t	1641626844
1	14	t	t	t	1641626844
1	15	t	t	t	1641626844
1	16	t	t	t	1641626844
1	17	t	t	t	1641626844
1	18	t	t	t	1641626844
1	19	t	t	t	1641626844
1	20	t	t	t	1641626844
1	21	t	t	t	1641626844
1	22	t	t	t	1641626844
1	23	t	t	t	1641626844
1	24	t	t	t	1641626844
1	25	t	t	t	1641626844
1	26	t	t	t	1641626844
1	27	t	t	t	1641626844
1	28	t	t	t	1641626844
1	29	t	t	t	1641626844
1	30	t	t	t	1641626844
1	31	t	t	t	1641626844
1	32	t	t	t	1641626844
1	33	t	t	t	1641626844
1	34	t	t	t	1641626844
1	35	t	t	t	1641626844
1	36	t	t	t	1641626844
1	37	t	t	t	1641626844
1	38	t	t	t	1641626844
1	39	t	t	t	1641626844
1	40	t	t	t	1641626844
1	41	t	t	t	1641626844
1	42	t	t	t	1641626844
1	43	t	t	t	1641626844
1	44	t	t	t	1641626844
1	45	t	t	t	1641626844
1	46	t	t	t	1641626844
1	47	t	t	t	1641626844
1	48	t	t	t	1641626844
1	49	t	t	t	1641626844
1	50	t	t	t	1641626844
1	51	t	t	t	1641626844
1	52	t	t	t	1641626844
1	53	t	t	t	1641626844
1	54	t	t	t	1641626844
1	55	t	t	t	1641626844
1	56	t	t	t	1641626844
1	57	t	t	t	1641626844
1	58	t	t	t	1641626844
1	59	t	t	t	1641626844
1	60	t	t	t	1641626844
1	115	t	t	t	1641630451
1	105	t	t	t	1641630452
\.


--
-- Data for Name: procs; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.procs (proc_id, proc_build_id, proc_pid, proc_ppid, proc_pgid, proc_name, proc_state, proc_error, proc_exit_code, proc_started, proc_stopped, proc_machine, proc_platform, proc_environ) FROM stdin;
1	1	1	0	1	drone	failure	Error response from daemon: manifest for woodpeckerci/plugin-git:test not found: manifest unknown: manifest unknown	1	1641630525	1641630527	PC-Maddl-HOME		{}\n
2	1	2	1	2	git	success		0	1641630525	1641630527	PC-Maddl-HOME		null\n
3	1	3	1	3	Print	skipped		0	0	0			null\n
\.


--
-- Data for Name: registry; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.registry (registry_id, registry_repo_id, registry_addr, registry_email, registry_username, registry_password, registry_token) FROM stdin;
\.


--
-- Data for Name: repos; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.repos (repo_id, repo_user_id, repo_owner, repo_name, repo_full_name, repo_avatar, repo_link, repo_clone, repo_branch, repo_timeout, repo_private, repo_trusted, repo_allow_pr, repo_allow_push, repo_allow_deploys, repo_allow_tags, repo_hash, repo_scm, repo_config_path, repo_gated, repo_visibility, repo_counter, repo_active, repo_fallback) FROM stdin;
1	0	test	a	test/a	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/a	http://10.40.8.5:3000/test/a.git	main	0	f	f	f	f	f	f		git		f		0	f	f
2	0	test	aa	test/aa	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/aa	http://10.40.8.5:3000/test/aa.git	main	0	f	f	f	f	f	f		git		f		0	f	f
3	0	test	aaaa	test/aaaa	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/aaaa	http://10.40.8.5:3000/test/aaaa.git	master	0	f	f	f	f	f	f		git		f		0	f	f
4	0	test	asciidoc-test	test/asciidoc-test	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/asciidoc-test	http://10.40.8.5:3000/test/asciidoc-test.git	master	0	f	f	f	f	f	f		git		f		0	f	f
5	0	test	bigLFS	test/bigLFS	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/bigLFS	http://10.40.8.5:3000/test/bigLFS.git	master	0	f	f	f	f	f	f		git		f		0	f	f
6	0	test	codeberg-gitea	test/codeberg-gitea	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/codeberg-gitea	http://10.40.8.5:3000/test/codeberg-gitea.git	codeberg-1.15	0	f	f	f	f	f	f		git		f		0	f	f
7	0	fnetX	codeberg-gitea	fnetX/codeberg-gitea	http://10.40.8.5:3000/avatars/2a635c272612fabab4fa1d11c9720d1a	http://10.40.8.5:3000/fnetX/codeberg-gitea	http://10.40.8.5:3000/fnetX/codeberg-gitea.git	pages	0	f	f	f	f	f	f		git		f		0	f	f
8	0	test	CSV-deom	test/CSV-deom	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/CSV-deom	http://10.40.8.5:3000/test/CSV-deom.git	master	0	f	f	f	f	f	f		git		f		0	f	f
9	0	test	empty	test/empty	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/empty	http://10.40.8.5:3000/test/empty.git	master	0	f	f	f	f	f	f		git		f		0	f	f
10	0	test3	fdsa	test3/fdsa	http://10.40.8.5:3000/avatar/68985079da908d72fcfca7b557d8f729	http://10.40.8.5:3000/test3/fdsa	http://10.40.8.5:3000/test3/fdsa.git	main	0	f	f	f	f	f	f		git		f		0	f	f
11	0	test	fdsa	test/fdsa	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/fdsa	http://10.40.8.5:3000/test/fdsa.git	main	0	t	f	f	f	f	f		git		f		0	f	f
12	0	test	fdsa-mig	test/fdsa-mig	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/fdsa-mig	http://10.40.8.5:3000/test/fdsa-mig.git	main	0	f	f	f	f	f	f		git		f		0	f	f
13	0	test	fdsa-mig2	test/fdsa-mig2	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/fdsa-mig2	http://10.40.8.5:3000/test/fdsa-mig2.git	main	0	f	f	f	f	f	f		git		f		0	f	f
14	0	test	fdsaddd	test/fdsaddd	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/fdsaddd	http://10.40.8.5:3000/test/fdsaddd.git	main	0	f	f	f	f	f	f		git		f		0	f	f
15	0	df	fdsafdsafdsa	df/fdsafdsafdsa	http://10.40.8.5:3000/avatars/eff7d5dba32b4da32d9a67a519434d3f	http://10.40.8.5:3000/df/fdsafdsafdsa	http://10.40.8.5:3000/df/fdsafdsafdsa.git	main	0	f	f	f	f	f	f		git		f		0	f	f
16	0	test	freebsd	test/freebsd	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/freebsd	http://10.40.8.5:3000/test/freebsd.git	main	0	f	f	f	f	f	f		git		f		0	f	f
17	0	test	FreeBSD_ports	test/FreeBSD_ports	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/FreeBSD_ports	http://10.40.8.5:3000/test/FreeBSD_ports.git	main	0	f	f	f	f	f	f		git		f		0	f	f
18	0	test	Gadgetbridge	test/Gadgetbridge	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/Gadgetbridge	http://10.40.8.5:3000/test/Gadgetbridge.git	master	0	f	f	f	f	f	f		git		f		0	f	f
19	0	test	gcc	test/gcc	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/gcc	http://10.40.8.5:3000/test/gcc.git	master	0	f	f	f	f	f	f		git		f		0	f	f
20	0	test	gitea	test/gitea	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/gitea	http://10.40.8.5:3000/test/gitea.git	main	0	f	f	f	f	f	f		git		f		0	f	f
21	0	test	github-orgmode-tests	test/github-orgmode-tests	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/github-orgmode-tests	http://10.40.8.5:3000/test/github-orgmode-tests.git	master	0	f	f	f	f	f	f		git		f		0	f	f
22	0	test	go-hexcolor	test/go-hexcolor	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/go-hexcolor	http://10.40.8.5:3000/test/go-hexcolor.git	master	0	f	f	f	f	f	f		git		f		0	f	f
23	0	test	go-sdk	test/go-sdk	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/go-sdk	http://10.40.8.5:3000/test/go-sdk.git	master	0	f	f	f	f	f	f		git		f		0	f	f
24	0	test	go-version	test/go-version	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/go-version	http://10.40.8.5:3000/test/go-version.git	master	0	f	f	f	f	f	f		git		f		0	f	f
25	0	test	go-version2	test/go-version2	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/go-version2	http://10.40.8.5:3000/test/go-version2.git	master	0	f	f	f	f	f	f		git		f		0	f	f
26	0	CI-Tests	helm-release	CI-Tests/helm-release	http://10.40.8.5:3000/avatars/999baa049c222f6d4d89f49018ecf687	http://10.40.8.5:3000/CI-Tests/helm-release	http://10.40.8.5:3000/CI-Tests/helm-release.git	master	0	f	f	f	f	f	f		git		f		0	f	f
27	0	test	init	test/init	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/init	http://10.40.8.5:3000/test/init.git	main	0	f	f	f	f	f	f		git		f		0	f	f
28	0	df	init	df/init	http://10.40.8.5:3000/avatars/eff7d5dba32b4da32d9a67a519434d3f	http://10.40.8.5:3000/df/init	http://10.40.8.5:3000/df/init.git	main	0	f	f	f	f	f	f		git		f		0	f	f
29	0	23r3e2	init	23r3e2/init	http://10.40.8.5:3000/avatars/9bc4c5e506b1bfbe3033e35f9e78428b	http://10.40.8.5:3000/23r3e2/init	http://10.40.8.5:3000/23r3e2/init.git	main	0	f	f	f	f	f	f		git		f		0	f	f
30	0	test	LFS-TEST	test/LFS-TEST	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/LFS-TEST	http://10.40.8.5:3000/test/LFS-TEST.git	main	0	f	f	f	f	f	f		git		f		0	f	f
31	0	test	mig-opendev-test	test/mig-opendev-test	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/mig-opendev-test	http://10.40.8.5:3000/test/mig-opendev-test.git	master	0	f	f	f	f	f	f		git		f		0	f	f
32	0	test	mig-opendev-test2	test/mig-opendev-test2	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/mig-opendev-test2	http://10.40.8.5:3000/test/mig-opendev-test2.git	master	0	f	f	f	f	f	f		git		f		0	f	f
33	0	test	mopidy-autoplay	test/mopidy-autoplay	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/mopidy-autoplay	http://10.40.8.5:3000/test/mopidy-autoplay.git	master	0	f	f	f	f	f	f		git		f		0	f	f
34	0	test	namla	test/namla	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/namla	http://10.40.8.5:3000/test/namla.git	master	0	f	f	f	f	f	f		git		f		0	f	f
36	0	CI-Tests	pages-zola	CI-Tests/pages-zola	http://10.40.8.5:3000/avatars/999baa049c222f6d4d89f49018ecf687	http://10.40.8.5:3000/CI-Tests/pages-zola	http://10.40.8.5:3000/CI-Tests/pages-zola.git	main	0	f	f	f	f	f	f		git		f		0	f	f
38	0	CI-Tests	plugin-settings	CI-Tests/plugin-settings	http://10.40.8.5:3000/avatars/999baa049c222f6d4d89f49018ecf687	http://10.40.8.5:3000/CI-Tests/plugin-settings	http://10.40.8.5:3000/CI-Tests/plugin-settings.git	main	0	f	f	f	f	f	f		git		f		0	f	f
40	0	test	produceit	test/produceit	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/produceit	http://10.40.8.5:3000/test/produceit.git	master	0	f	f	f	f	f	f		git		f		0	f	f
42	0	test	Remmina-pull	test/Remmina-pull	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/Remmina-pull	http://10.40.8.5:3000/test/Remmina-pull.git	main	0	f	f	f	f	f	f		git		f		0	f	f
44	0	CI-Tests	settings	CI-Tests/settings	http://10.40.8.5:3000/avatars/999baa049c222f6d4d89f49018ecf687	http://10.40.8.5:3000/CI-Tests/settings	http://10.40.8.5:3000/CI-Tests/settings.git	master	0	f	f	f	f	f	f		git		f		0	f	f
46	0	581	tag-issue	581/tag-issue	http://10.40.8.5:3000/avatars/c6e19e830859f2cb9f7c8f8cacb8d2a6	http://10.40.8.5:3000/581/tag-issue	http://10.40.8.5:3000/581/tag-issue.git	main	0	f	f	f	f	f	f		git		f		0	f	f
48	0	23r3e2	test-clone	23r3e2/test-clone	http://10.40.8.5:3000/avatars/9bc4c5e506b1bfbe3033e35f9e78428b	http://10.40.8.5:3000/23r3e2/test-clone	http://10.40.8.5:3000/23r3e2/test-clone.git	master	0	t	f	f	f	f	f		git		f		0	f	f
50	0	test	test-gitea-migration-release-draft	test/test-gitea-migration-release-draft	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/test-gitea-migration-release-draft	http://10.40.8.5:3000/test/test-gitea-migration-release-draft.git	main	0	f	f	f	f	f	f		git		f		0	f	f
52	0	test	testCIservices	test/testCIservices	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/testCIservices	http://10.40.8.5:3000/test/testCIservices.git	master	0	f	f	f	f	f	f		git		f		0	f	f
54	0	test	testStrangeCommits	test/testStrangeCommits	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/testStrangeCommits	http://10.40.8.5:3000/test/testStrangeCommits.git	master	0	f	f	f	f	f	f		git		f		0	f	f
56	0	test	vim	test/vim	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/vim	http://10.40.8.5:3000/test/vim.git	master	0	f	f	f	f	f	f		git		f		0	f	f
58	0	test	ww	test/ww	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/ww	http://10.40.8.5:3000/test/ww.git	main	0	f	f	f	f	f	f		git		f		0	f	f
60	0	test	x_bows	test/x_bows	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/x_bows	http://10.40.8.5:3000/test/x_bows.git	master	0	f	f	f	f	f	f		git		f		0	f	f
35	0	orga	oio	orga/oio	http://10.40.8.5:3000/avatars/93778f25b68b74ce5d69b8f8634bbf36	http://10.40.8.5:3000/orga/oio	http://10.40.8.5:3000/orga/oio.git	main	0	f	f	f	f	f	f		git		f		0	f	f
37	0	test	pathological	test/pathological	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/pathological	http://10.40.8.5:3000/test/pathological.git	master	0	f	f	f	f	f	f		git		f		0	f	f
39	0	test	PNGs	test/PNGs	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/PNGs	http://10.40.8.5:3000/test/PNGs.git	main	0	f	f	f	f	f	f		git		f		0	f	f
41	0	test	pyrocko	test/pyrocko	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/pyrocko	http://10.40.8.5:3000/test/pyrocko.git	master	0	f	f	f	f	f	f		git		f		0	f	f
43	0	test	reStructuredText_ReST	test/reStructuredText_ReST	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/reStructuredText_ReST	http://10.40.8.5:3000/test/reStructuredText_ReST.git	master	0	f	f	f	f	f	f		git		f		0	f	f
45	0	df	spam	df/spam	http://10.40.8.5:3000/avatars/eff7d5dba32b4da32d9a67a519434d3f	http://10.40.8.5:3000/df/spam	http://10.40.8.5:3000/df/spam.git	main	0	f	f	f	f	f	f		git		f		0	f	f
47	0	test	tea	test/tea	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/tea	http://10.40.8.5:3000/test/tea.git	master	0	f	f	f	f	f	f		git		f		0	f	f
49	0	test	test-event	test/test-event	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/test-event	http://10.40.8.5:3000/test/test-event.git	master	0	f	f	f	f	f	f		git		f		0	f	f
51	0	test	testCI	test/testCI	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/testCI	http://10.40.8.5:3000/test/testCI.git	master	0	t	f	f	f	f	f		git		f		0	f	f
53	0	hahaO	testCIservices	hahaO/testCIservices	http://10.40.8.5:3000/avatars/2a6eec168901fffe947ddd5a69dbdb82	http://10.40.8.5:3000/hahaO/testCIservices	http://10.40.8.5:3000/hahaO/testCIservices.git	master	0	f	f	f	f	f	f		git		f		0	f	f
55	0	CI-Tests	version-test	CI-Tests/version-test	http://10.40.8.5:3000/avatars/999baa049c222f6d4d89f49018ecf687	http://10.40.8.5:3000/CI-Tests/version-test	http://10.40.8.5:3000/CI-Tests/version-test.git	master	0	f	f	f	f	f	f		git		f		0	f	f
57	0	test	woodpecker	test/woodpecker	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/woodpecker	http://10.40.8.5:3000/test/woodpecker.git	master	0	f	f	f	f	f	f		git		f		0	f	f
59	0	test	xss-issue	test/xss-issue	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	http://10.40.8.5:3000/test/xss-issue	http://10.40.8.5:3000/test/xss-issue.git	master	0	f	f	f	f	f	f		git		f		0	f	f
115	1	2	testCIservices	2/testCIservices	http://10.40.8.5:3000/avatars/c81e728d9d4c2f636f067f89cc14862c	http://10.40.8.5:3000/2/testCIservices	http://10.40.8.5:3000/2/testCIservices.git	master	60	f	f	t	t	t	t	FOUXTSNL2GXK7JP2SQQJVWVAS6J4E4SGIQYPAHEJBIFPVR46LLDA====	git	.drone.yml	f	public	0	t	t
105	1	2	settings	2/settings	http://10.40.8.5:3000/avatars/c81e728d9d4c2f636f067f89cc14862c	http://10.40.8.5:3000/2/settings	http://10.40.8.5:3000/2/settings.git	master	60	f	f	t	t	t	t	3OQA7X5CNGPTILDYLQSJFDML6U2W7UUFBPPP2G2LRBG3WETAYZLA====	git	.drone.yml	f	public	1	t	t
\.


--
-- Data for Name: secrets; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.secrets (secret_id, secret_repo_id, secret_name, secret_value, secret_images, secret_events, secret_skip_verify, secret_conceal) FROM stdin;
1	105	wow	\\x74657374	null\n	["push","tag","deployment","pull_request"]\n	f	f
2	105	n	\\x6e	null\n	["deployment"]\n	f	f
3	105	abc	\\x656466	null\n	["push"]\n	f	f
4	105	quak	\\x66647361	null\n	["pull-request"]\n	f	f
\.


--
-- Data for Name: senders; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.senders (sender_id, sender_repo_id, sender_login, sender_allow, sender_block) FROM stdin;
\.


--
-- Data for Name: tasks; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.tasks (task_id, task_data, task_labels, task_dependencies, task_run_on) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: root
--

COPY public.users (user_id, user_login, user_token, user_secret, user_expiry, user_email, user_avatar, user_active, user_admin, user_hash, user_synced) FROM stdin;
1	test	eyJhbGciOiJSUzI1NiIsImtpZCI6IldmbUJ1c2Q0RndUVWRmMjc2NHowUWlEYlJ3TnRBcU5pNVlXS1U1c2k0eEEiLCJ0eXAiOiJKV1QifQ.eyJnbnQiOjEsInR0IjowLCJleHAiOjE2NDE2MzQxMjcsImlhdCI6MTY0MTYzMDUyN30.Fu0wUP-08NpPjq737y6HOeyKN_-_SE4iOZr5yrH7S8Jrz8nIuNKfU7AvlypeMSJ7wo8e3cSTadbSH1polZuFv-Nb1AqWDDXeuXudm61BkF96sTslbSHd0nF7cOy6hqCfIAfQLQpqZTJZ4E26oOSSJxPfOOntOWhlEejRl5F-flXAoYAQLegHxdn9IfYJeM1eanZqF4k6dT9hthFp9v4fmUjODPPfHip_iS7ckPonP1E4-8KeNkU3O-lIS1fgrsbCDA8531FXIGB0U7cSur7H0picKGL6WSzAErPGntlNlQWYB5JedDtLN9Ionxy1Y9LKQON76XYL4gM1Ji98RCEXggVqd7TW0B1fGV-Jve2hU3fKaDyQywsCJp36mpnVaqb5eiTssncHixAwZE0C4yh_XsTd-WoVhsbqlEuDfPTjrtAK94mSzHJTcO3fbtE9L-MoPevQIPM7Yog0i2Xn1oPUCDXVXsV2yJriBiI_r2xbG0nz5Bwn8KAFZ0dNGJ7T9urqKaKMh9guE4jgYLIpRpod_Fd13_GAK0ebgF2CZJdjJT7eEGhzzcg4uFpFdIXL2kNgVN1D6YLMPw3HhVg7_MIfASbJgpcppFhYa4Fk-OpchL5-e_mMyeWogvaJA2wSpyY1f5zJlBnFuIyk_OdV0TwQ3b_TjutehsiibT9WRpOK8h8	eyJhbGciOiJSUzI1NiIsImtpZCI6IldmbUJ1c2Q0RndUVWRmMjc2NHowUWlEYlJ3TnRBcU5pNVlXS1U1c2k0eEEiLCJ0eXAiOiJKV1QifQ.eyJnbnQiOjEsInR0IjoxLCJleHAiOjE2NDQyNTg1MjcsImlhdCI6MTY0MTYzMDUyN30.iVtIGQ6VTgRI8L3xFD_YNvVBGZ6kdFb3ERdyOCIHC_CHhOEpZxVGawMGnNNooqbNdmOqJQ0RLJyiAirEKdxSVrtWvqub6uVMjjpeBylE1sAFymCGNJQf77dKvgPHW3QY5FvOSoOoNcRU2g99Bx8sbZhiI12GnNOB-abazrzICpOUikiTdb2ri3w_TNF2Ibrn-itSa1yuhmTrVpqXt_CT4MEfteiDmgjyqonmk-J_BqbcriF3DKAvrXNK1VKVU7xODcFSIRizlgA2kDmnpMT3Oo-Z1I37TFIGAuDOTgcceOPa7rXg_Mfd_jhL7bSH1BI4RsK0rgde3NaCQlU2n7yVOYGbJCSsSWwSAi-gCjjuTTPnQWe3ep3IWrB73_7tKG2_x7YxZ1nQCSFKouA5rZH4g6yoV8wdJh8_bX2Z64-MJBUl8E7JGM2urA5GY1abo0GZ6ZuQi2JS5WnG1iTL9pFlmOoTpN1DKtNE2PUE90GJwi0qGeACif9uJBXQPDAgKk7fbUxKYQobc6ko2CJ1isoRjbi8-GsJ9lhw7tXno5zfAvN3eps9SYgmIRNh0t_vx-LMBezSTSEcTJpv-7Ap6F10GD3E9KmGcYrOMvdtaYgkWFXO6rh49uElUVid-C1tNVpKjnj7ewUosQo9MHSn-d5l1df0rJSueXcaUMSqRSrEzqQ	1641634127	test@test.test	http://10.40.8.5:3000/avatars/d6c72f5d7e2a070b52e1194969df2cfe	f	f	OBW2OF5QH3NMCYJ44VU5B5YEQ5LHZLTFW2FDSAJ4R4JVZ4HWSNVQ====	1641630445
\.


--
-- Name: agents_agent_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('public.agents_agent_id_seq', 1, false);


--
-- Name: builds_build_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('public.builds_build_id_seq', 1, true);


--
-- Name: config_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('public.config_config_id_seq', 1, true);


--
-- Name: files_file_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('public.files_file_id_seq', 1, false);


--
-- Name: logs_log_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('public.logs_log_id_seq', 1, false);


--
-- Name: procs_proc_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('public.procs_proc_id_seq', 3, true);


--
-- Name: registry_registry_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('public.registry_registry_id_seq', 1, false);


--
-- Name: repos_repo_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('public.repos_repo_id_seq', 122, true);


--
-- Name: secrets_secret_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('public.secrets_secret_id_seq', 4, true);


--
-- Name: senders_sender_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('public.senders_sender_id_seq', 1, false);


--
-- Name: users_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: root
--

SELECT pg_catalog.setval('public.users_user_id_seq', 1, true);


--
-- Name: agents agents_agent_addr_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_agent_addr_key UNIQUE (agent_addr);


--
-- Name: agents agents_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_pkey PRIMARY KEY (agent_id);


--
-- Name: build_config build_config_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.build_config
    ADD CONSTRAINT build_config_pkey PRIMARY KEY (config_id, build_id);


--
-- Name: builds builds_build_number_build_repo_id_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.builds
    ADD CONSTRAINT builds_build_number_build_repo_id_key UNIQUE (build_number, build_repo_id);


--
-- Name: builds builds_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.builds
    ADD CONSTRAINT builds_pkey PRIMARY KEY (build_id);


--
-- Name: config config_config_hash_config_repo_id_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.config
    ADD CONSTRAINT config_config_hash_config_repo_id_key UNIQUE (config_hash, config_repo_id);


--
-- Name: config config_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.config
    ADD CONSTRAINT config_pkey PRIMARY KEY (config_id);


--
-- Name: files files_file_proc_id_file_name_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_file_proc_id_file_name_key UNIQUE (file_proc_id, file_name);


--
-- Name: files files_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (file_id);


--
-- Name: logs logs_log_job_id_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_log_job_id_key UNIQUE (log_job_id);


--
-- Name: logs logs_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_pkey PRIMARY KEY (log_id);


--
-- Name: migrations migrations_name_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT migrations_name_key UNIQUE (name);


--
-- Name: perms perms_perm_user_id_perm_repo_id_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.perms
    ADD CONSTRAINT perms_perm_user_id_perm_repo_id_key UNIQUE (perm_user_id, perm_repo_id);


--
-- Name: procs procs_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.procs
    ADD CONSTRAINT procs_pkey PRIMARY KEY (proc_id);


--
-- Name: procs procs_proc_build_id_proc_pid_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.procs
    ADD CONSTRAINT procs_proc_build_id_proc_pid_key UNIQUE (proc_build_id, proc_pid);


--
-- Name: registry registry_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.registry
    ADD CONSTRAINT registry_pkey PRIMARY KEY (registry_id);


--
-- Name: registry registry_registry_addr_registry_repo_id_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.registry
    ADD CONSTRAINT registry_registry_addr_registry_repo_id_key UNIQUE (registry_addr, registry_repo_id);


--
-- Name: repos repos_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.repos
    ADD CONSTRAINT repos_pkey PRIMARY KEY (repo_id);


--
-- Name: repos repos_repo_full_name_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.repos
    ADD CONSTRAINT repos_repo_full_name_key UNIQUE (repo_full_name);


--
-- Name: secrets secrets_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.secrets
    ADD CONSTRAINT secrets_pkey PRIMARY KEY (secret_id);


--
-- Name: secrets secrets_secret_name_secret_repo_id_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.secrets
    ADD CONSTRAINT secrets_secret_name_secret_repo_id_key UNIQUE (secret_name, secret_repo_id);


--
-- Name: senders senders_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.senders
    ADD CONSTRAINT senders_pkey PRIMARY KEY (sender_id);


--
-- Name: senders senders_sender_repo_id_sender_login_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.senders
    ADD CONSTRAINT senders_sender_repo_id_sender_login_key UNIQUE (sender_repo_id, sender_login);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (task_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- Name: users users_user_login_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_user_login_key UNIQUE (user_login);


--
-- Name: file_build_ix; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX file_build_ix ON public.files USING btree (file_build_id);


--
-- Name: file_proc_ix; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX file_proc_ix ON public.files USING btree (file_proc_id);


--
-- Name: ix_build_author; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX ix_build_author ON public.builds USING btree (build_author);


--
-- Name: ix_build_repo; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX ix_build_repo ON public.builds USING btree (build_repo_id);


--
-- Name: ix_perms_repo; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX ix_perms_repo ON public.perms USING btree (perm_repo_id);


--
-- Name: ix_perms_user; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX ix_perms_user ON public.perms USING btree (perm_user_id);


--
-- Name: ix_registry_repo; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX ix_registry_repo ON public.registry USING btree (registry_repo_id);


--
-- Name: ix_secrets_repo; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX ix_secrets_repo ON public.secrets USING btree (secret_repo_id);


--
-- Name: proc_build_ix; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX proc_build_ix ON public.procs USING btree (proc_build_id);


--
-- Name: sender_repo_ix; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX sender_repo_ix ON public.senders USING btree (sender_repo_id);


--
-- Name: build_config build_config_build_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.build_config
    ADD CONSTRAINT build_config_build_id_fkey FOREIGN KEY (build_id) REFERENCES public.builds(build_id);


--
-- Name: build_config build_config_config_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.build_config
    ADD CONSTRAINT build_config_config_id_fkey FOREIGN KEY (config_id) REFERENCES public.config(config_id);


--
-- PostgreSQL database dump complete
--

