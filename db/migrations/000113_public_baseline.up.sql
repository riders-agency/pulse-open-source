--
-- PostgreSQL database dump
--


-- Dumped from database version 16.15 (Ubuntu 16.15-0ubuntu0.24.04.1)
-- Dumped by pg_dump version 16.15 (Ubuntu 16.15-0ubuntu0.24.04.1)

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
-- Name: ai_agent_checkpoint; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA ai_agent_checkpoint;


--
-- Name: SCHEMA ai_agent_checkpoint; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA ai_agent_checkpoint IS 'Restricted LangGraph checkpoint/store schema for pulse-ai-agent. The AI role must not receive grants on Riders Pulse business tables.';


--
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: unaccent; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS unaccent WITH SCHEMA public;


--
-- Name: EXTENSION unaccent; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION unaccent IS 'text search dictionary that removes accents';


--
-- Name: auth_identity_versions_on_delete(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.auth_identity_versions_on_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_version_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_version_no
    FROM auth_identity_versions
   WHERE auth_identity_id = OLD.auth_identity_id;

  INSERT INTO auth_identity_versions (
    auth_identity_id,
    version_no,
    operation_type,
    snapshot,
    changed_by_user_id,
    request_id
  ) VALUES (
    OLD.auth_identity_id,
    v_version_no,
    'delete',
    to_jsonb(OLD),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN OLD;
END;
$$;


--
-- Name: auth_identity_versions_on_write(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.auth_identity_versions_on_write() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_version_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_version_no
    FROM auth_identity_versions
   WHERE auth_identity_id = NEW.auth_identity_id;

  INSERT INTO auth_identity_versions (
    auth_identity_id,
    version_no,
    operation_type,
    snapshot,
    changed_by_user_id,
    request_id
  ) VALUES (
    NEW.auth_identity_id,
    v_version_no,
    lower(TG_OP),
    to_jsonb(NEW),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN NEW;
END;
$$;


--
-- Name: company_membership_versions_on_delete(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.company_membership_versions_on_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_version_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_version_no
    FROM company_membership_versions
   WHERE company_id = OLD.company_id
     AND user_id = OLD.user_id;

  INSERT INTO company_membership_versions (
    company_id,
    user_id,
    version_no,
    operation_type,
    snapshot,
    changed_by_user_id,
    request_id
  ) VALUES (
    OLD.company_id,
    OLD.user_id,
    v_version_no,
    'delete',
    to_jsonb(OLD),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN OLD;
END;
$$;


--
-- Name: company_membership_versions_on_write(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.company_membership_versions_on_write() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_version_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_version_no
    FROM company_membership_versions
   WHERE company_id = NEW.company_id
     AND user_id = NEW.user_id;

  INSERT INTO company_membership_versions (
    company_id,
    user_id,
    version_no,
    operation_type,
    snapshot,
    changed_by_user_id,
    request_id
  ) VALUES (
    NEW.company_id,
    NEW.user_id,
    v_version_no,
    lower(TG_OP),
    to_jsonb(NEW),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN NEW;
END;
$$;


--
-- Name: company_versions_on_delete(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.company_versions_on_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_version_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_version_no
    FROM company_versions
   WHERE company_id = OLD.company_id;

  INSERT INTO company_versions (
    company_id,
    version_no,
    operation_type,
    snapshot,
    changed_by_user_id,
    request_id
  ) VALUES (
    OLD.company_id,
    v_version_no,
    'delete',
    to_jsonb(OLD),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN OLD;
END;
$$;


--
-- Name: company_versions_on_write(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.company_versions_on_write() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_version_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_version_no
    FROM company_versions
   WHERE company_id = NEW.company_id;

  INSERT INTO company_versions (
    company_id,
    version_no,
    operation_type,
    snapshot,
    changed_by_user_id,
    request_id
  ) VALUES (
    NEW.company_id,
    v_version_no,
    lower(TG_OP),
    to_jsonb(NEW),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN NEW;
END;
$$;


--
-- Name: device_credential_versions_on_delete(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.device_credential_versions_on_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_version_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_version_no
    FROM device_credential_versions
   WHERE device_credential_id = OLD.device_credential_id;

  INSERT INTO device_credential_versions (
    device_credential_id,
    version_no,
    operation_type,
    snapshot,
    changed_by_user_id,
    request_id
  ) VALUES (
    OLD.device_credential_id,
    v_version_no,
    'delete',
    to_jsonb(OLD),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN OLD;
END;
$$;


--
-- Name: device_credential_versions_on_write(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.device_credential_versions_on_write() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_version_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_version_no
    FROM device_credential_versions
   WHERE device_credential_id = NEW.device_credential_id;

  INSERT INTO device_credential_versions (
    device_credential_id,
    version_no,
    operation_type,
    snapshot,
    changed_by_user_id,
    request_id
  ) VALUES (
    NEW.device_credential_id,
    v_version_no,
    lower(TG_OP),
    to_jsonb(NEW),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN NEW;
END;
$$;


--
-- Name: enrich_payment_current_notification_input(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enrich_payment_current_notification_input() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  cp jsonb;
  pulse jsonb;
  client_name text;
  service_name text;
  amount_label text;
  period_label text;
  next_label text;
  source_label text;
  method_label text;
  client_url text;
  service_url text;
  payment_url text;
  telegram_block text;
  max_block text;
  email_block text;
  subject_prefix text;
BEGIN
  NEW.input_json := COALESCE(NEW.input_json, '{}'::jsonb);
  cp := NEW.input_json -> 'current_payment';
  IF NEW.source_type <> 'module_trigger' OR cp IS NULL OR cp = '{}'::jsonb OR cp = 'null'::jsonb THEN
    RETURN NEW;
  END IF;

  client_name := COALESCE(NULLIF(cp ->> 'client_name', ''), 'Клиент');
  service_name := COALESCE(NULLIF(cp ->> 'service_name', ''), NULLIF(cp ->> 'service_summary', ''), 'Услуга');
  amount_label := pulse_payment_money_label(cp ->> 'amount', cp ->> 'currency');
  period_label := COALESCE(NULLIF(cp ->> 'period_label', ''), pulse_payment_date_label(cp ->> 'period_start'));
  next_label := pulse_payment_date_label(COALESCE(NULLIF(cp ->> 'next_payment_due_date', ''), NULLIF(cp ->> 'next_period_start', '')));
  source_label := COALESCE(NULLIF(cp ->> 'source_label', ''), 'ручное внесение');
  method_label := COALESCE(NULLIF(cp ->> 'payment_method_label', ''), 'по счету с юрлица');
  client_url := COALESCE(NULLIF(cp #>> '{links,client_link,url}', ''), NULLIF(cp #>> '{links,client_link,href}', ''));
  service_url := COALESCE(NULLIF(cp #>> '{links,client_service_link,url}', ''), NULLIF(cp #>> '{links,client_service_link,href}', ''));
  payment_url := COALESCE(NULLIF(cp #>> '{links,payment_link,url}', ''), NULLIF(cp #>> '{links,payment_link,href}', ''));

  telegram_block :=
    '<b>💰 Поступила оплата</b>' || chr(10) || chr(10) ||
    '<b>' || pulse_html_escape(client_name) || '</b> оплатил <b>' || pulse_html_escape(amount_label) || '</b>' || chr(10) ||
    'Услуга: ' || pulse_html_escape(service_name) || chr(10) ||
    'Период: ' || pulse_html_escape(period_label) || chr(10) ||
    'Следующий платеж: ' || pulse_html_escape(next_label) || chr(10) ||
    'Способ: ' || pulse_html_escape(method_label) || chr(10) ||
    'Источник: ' || pulse_html_escape(source_label) ||
    CASE WHEN client_url <> '' THEN chr(10) || chr(10) || 'Клиент: <a href="' || pulse_html_escape(client_url) || '">' || pulse_html_escape(client_name) || '</a>' ELSE '' END ||
    CASE WHEN service_url <> '' THEN chr(10) || 'Услуга: <a href="' || pulse_html_escape(service_url) || '">' || pulse_html_escape(service_name) || '</a>' ELSE '' END ||
    CASE WHEN payment_url <> '' THEN chr(10) || 'Оплата: <a href="' || pulse_html_escape(payment_url) || '">' || pulse_html_escape(amount_label || ' за ' || period_label) || '</a>' ELSE '' END ||
    CASE WHEN service_url <> '' THEN chr(10) || 'Следующая дата: <a href="' || pulse_html_escape(service_url) || '">изменить</a>' ELSE '' END ||
    chr(10) || chr(10);

  max_block :=
    '💰 Поступила оплата' || chr(10) || chr(10) ||
    client_name || ' оплатил ' || amount_label || chr(10) ||
    'Услуга: ' || service_name || chr(10) ||
    'Период: ' || period_label || chr(10) ||
    'Следующий платеж: ' || next_label || chr(10) ||
    'Способ: ' || method_label || chr(10) ||
    'Источник: ' || source_label ||
    CASE WHEN client_url <> '' THEN chr(10) || chr(10) || 'Клиент: ' || client_url ELSE '' END ||
    CASE WHEN service_url <> '' THEN chr(10) || 'Услуга: ' || service_url ELSE '' END ||
    CASE WHEN payment_url <> '' THEN chr(10) || 'Оплата: ' || payment_url ELSE '' END ||
    chr(10) || chr(10);

  email_block := '<tr><td style="padding:18px 24px;border-bottom:1px solid #252a36;background:#172033;">' ||
    '<div style="font-size:13px;color:#7dd3fc;text-transform:uppercase;letter-spacing:.06em;margin-bottom:8px;">Поступила оплата</div>' ||
    '<div style="color:#e8ecf4;font-size:15px;line-height:1.55;">' ||
    '<div><strong>Клиент:</strong> ' || pulse_html_escape(client_name) || '</div>' ||
    '<div><strong>Услуга:</strong> ' || pulse_html_escape(service_name) || '</div>' ||
    '<div><strong>Сумма:</strong> ' || pulse_html_escape(amount_label) || '</div>' ||
    '<div><strong>Период:</strong> ' || pulse_html_escape(period_label) || '</div>' ||
    '<div><strong>Следующий платеж:</strong> ' || pulse_html_escape(next_label) || '</div>' ||
    '<div><strong>Способ оплаты:</strong> ' || pulse_html_escape(method_label) || '</div>' ||
    '<div><strong>Источник события:</strong> ' || pulse_html_escape(source_label) || '</div>' ||
    '</div><div style="margin-top:12px;">' ||
    CASE WHEN client_url <> '' THEN '<a href="' || pulse_html_escape(client_url) || '" style="display:inline-block;margin-right:8px;color:#6ea8fe;text-decoration:none;font-weight:600;">Открыть клиента</a>' ELSE '' END ||
    CASE WHEN service_url <> '' THEN '<a href="' || pulse_html_escape(service_url) || '" style="display:inline-block;margin-right:8px;color:#6ea8fe;text-decoration:none;font-weight:600;">Открыть услугу</a>' ELSE '' END ||
    CASE WHEN payment_url <> '' THEN '<a href="' || pulse_html_escape(payment_url) || '" style="display:inline-block;margin-right:8px;color:#6ea8fe;text-decoration:none;font-weight:600;">Открыть оплату</a>' ELSE '' END ||
    '</div></td></tr>';

  subject_prefix := client_name || ' оплатил ' || amount_label || ' · ';
  pulse := COALESCE(NEW.input_json -> '_pulse', '{}'::jsonb) ||
    jsonb_build_object('notification_current_payment_pre_rendered', true);

  NEW.input_json := jsonb_set(NEW.input_json, '{_pulse}', pulse, true) ||
    jsonb_build_object(
      'telegram_current_payment_block', telegram_block,
      'max_current_payment_block', max_block,
      'email_current_payment_block', email_block,
      'email_current_payment_subject', subject_prefix
    );
  RETURN NEW;
END $$;


--
-- Name: integration_doc_chunks_set_search_vector(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.integration_doc_chunks_set_search_vector() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('simple', unaccent(coalesce(NEW.heading, ''))), 'A') ||
    setweight(to_tsvector('simple', unaccent(coalesce(NEW.content, ''))), 'B');
  RETURN NEW;
END;
$$;


--
-- Name: integration_docs_set_search_vector(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.integration_docs_set_search_vector() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('simple', unaccent(coalesce(NEW.doc_key, ''))), 'A') ||
    setweight(to_tsvector('simple', unaccent(coalesce(NEW.title, ''))), 'A') ||
    setweight(to_tsvector('simple', unaccent(coalesce(NEW.summary, ''))), 'B') ||
    setweight(to_tsvector('simple', unaccent(coalesce(NEW.provider_key, ''))), 'B') ||
    setweight(to_tsvector('simple', unaccent(coalesce(array_to_string(NEW.action_type_keys, ' '), ''))), 'A') ||
    setweight(to_tsvector('simple', unaccent(coalesce(array_to_string(NEW.dependency_type_keys, ' '), ''))), 'B') ||
    setweight(to_tsvector('simple', unaccent(coalesce(array_to_string(NEW.tags, ' '), ''))), 'B') ||
    setweight(to_tsvector('simple', unaccent(coalesce(array_to_string(NEW.intents, ' '), ''))), 'A') ||
    setweight(to_tsvector('simple', unaccent(coalesce(NEW.markdown, ''))), 'C');
  NEW.updated_at := now();
  RETURN NEW;
END;
$$;


--
-- Name: project_member_versions_on_delete(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.project_member_versions_on_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_no
    FROM project_member_versions
   WHERE project_id = OLD.project_id AND user_id = OLD.user_id;

  INSERT INTO project_member_versions (
    project_id, user_id, version_no, operation_type, snapshot,
    changed_by_user_id, request_id
  ) VALUES (
    OLD.project_id,
    OLD.user_id,
    v_no,
    'delete'::text,
    to_jsonb(OLD),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN OLD;
END;
$$;


--
-- Name: project_member_versions_on_write(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.project_member_versions_on_write() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_no
    FROM project_member_versions
   WHERE project_id = NEW.project_id AND user_id = NEW.user_id;

  INSERT INTO project_member_versions (
    project_id, user_id, version_no, operation_type, snapshot,
    changed_by_user_id, request_id
  ) VALUES (
    NEW.project_id,
    NEW.user_id,
    v_no,
    lower(TG_OP)::text,
    to_jsonb(NEW),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN NEW;
END;
$$;


--
-- Name: project_versions_on_delete(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.project_versions_on_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_no
    FROM project_versions
   WHERE project_id = OLD.project_id;

  INSERT INTO project_versions (
    project_id, version_no, operation_type, snapshot,
    changed_by_user_id, request_id
  ) VALUES (
    OLD.project_id,
    v_no,
    'delete'::text,
    to_jsonb(OLD),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN OLD;
END;
$$;


--
-- Name: project_versions_on_write(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.project_versions_on_write() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_no
    FROM project_versions
   WHERE project_id = NEW.project_id;

  INSERT INTO project_versions (
    project_id, version_no, operation_type, snapshot,
    changed_by_user_id, request_id
  ) VALUES (
    NEW.project_id,
    v_no,
    lower(TG_OP)::text,
    to_jsonb(NEW),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN NEW;
END;
$$;


--
-- Name: pulse_html_escape(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.pulse_html_escape(value text) RETURNS text
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT replace(replace(replace(COALESCE($1, ''), '&', '&amp;'), '<', '&lt;'), '>', '&gt;')
$_$;


--
-- Name: pulse_payment_date_label(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.pulse_payment_date_label(value text) RETURNS text
    LANGUAGE plpgsql IMMUTABLE
    AS $_$
BEGIN
  IF COALESCE(value, '') ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}$' THEN
    RETURN to_char(value::date, 'DD.MM.YYYY');
  END IF;
  RETURN COALESCE(NULLIF(value, ''), '—');
END $_$;


--
-- Name: pulse_payment_money_label(text, text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.pulse_payment_money_label(amount_text text, currency_text text) RETURNS text
    LANGUAGE plpgsql IMMUTABLE
    AS $$
DECLARE
  amount_value numeric := 0;
  amount_label text;
  currency_label text;
BEGIN
  BEGIN
    amount_value := COALESCE(NULLIF(amount_text, '')::numeric, 0);
  EXCEPTION WHEN invalid_text_representation THEN
    amount_value := 0;
  END;
  amount_label := replace(trim(to_char(amount_value, 'FM999G999G999G999G990')), ',', ' ');
  currency_label := CASE COALESCE(NULLIF(currency_text, ''), 'RUB')
    WHEN 'RUB' THEN '₽'
    ELSE COALESCE(NULLIF(currency_text, ''), 'RUB')
  END;
  RETURN amount_label || ' ' || currency_label;
END $$;


--
-- Name: rp_current_actor(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.rp_current_actor() RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
  v text;
BEGIN
  v := current_setting('app.current_user_id', true);
  IF v IS NULL OR v = '' THEN
    RETURN NULL;
  END IF;
  RETURN v::uuid;
EXCEPTION
  WHEN others THEN
    RETURN NULL;
END;
$$;


--
-- Name: rp_current_request_id(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.rp_current_request_id() RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
  v text;
BEGIN
  v := current_setting('app.current_request_id', true);
  IF v IS NULL OR v = '' THEN
    RETURN NULL;
  END IF;
  RETURN v;
EXCEPTION
  WHEN others THEN
    RETURN NULL;
END;
$$;


--
-- Name: set_clients_search_vector(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_clients_search_vector() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.search_vector := to_tsvector(
    'russian'::regconfig,
    COALESCE(NEW.name, '') || ' ' ||
    COALESCE(NEW.short_name, '') || ' ' ||
    COALESCE(NEW.full_name, '') || ' ' ||
    COALESCE(NEW.inn, '') || ' ' ||
    COALESCE(NEW.kpp, '') || ' ' ||
    COALESCE(NEW.ogrn, '') || ' ' ||
    COALESCE(NEW.ogrnip, '') || ' ' ||
    COALESCE(NEW.comment, '')
  );
  RETURN NEW;
END;
$$;


--
-- Name: set_tasks_search_vector(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_tasks_search_vector() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.search_vector := to_tsvector(
    'russian'::regconfig,
    COALESCE(NEW.task_key, '') || ' ' ||
    COALESCE(NEW.title, '') || ' ' ||
    COALESCE(NEW.description, '') || ' ' ||
    COALESCE(NEW.labels::text, '')
  );
  RETURN NEW;
END;
$$;


--
-- Name: set_updated_at_now(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_updated_at_now() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$;


--
-- Name: user_versions_on_delete(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.user_versions_on_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_version_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_version_no
    FROM user_versions
   WHERE user_id = OLD.user_id;

  INSERT INTO user_versions (
    user_id,
    version_no,
    operation_type,
    snapshot,
    changed_by_user_id,
    request_id
  ) VALUES (
    OLD.user_id,
    v_version_no,
    'delete',
    to_jsonb(OLD),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN OLD;
END;
$$;


--
-- Name: user_versions_on_write(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.user_versions_on_write() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  v_version_no integer;
BEGIN
  SELECT COALESCE(MAX(version_no), 0) + 1
    INTO v_version_no
    FROM user_versions
   WHERE user_id = NEW.user_id;

  INSERT INTO user_versions (
    user_id,
    version_no,
    operation_type,
    snapshot,
    changed_by_user_id,
    request_id
  ) VALUES (
    NEW.user_id,
    v_version_no,
    lower(TG_OP),
    to_jsonb(NEW),
    rp_current_actor(),
    rp_current_request_id()
  );

  RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: ai_thread_scopes; Type: TABLE; Schema: ai_agent_checkpoint; Owner: -
--

CREATE TABLE ai_agent_checkpoint.ai_thread_scopes (
    checkpoint_namespace text NOT NULL,
    checkpoint_thread_id text NOT NULL,
    public_thread_id text NOT NULL,
    company_id text NOT NULL,
    project_id text NOT NULL,
    session_id text NOT NULL,
    state_schema_version integer NOT NULL,
    status text NOT NULL,
    first_seen_at timestamp with time zone DEFAULT now() NOT NULL,
    last_seen_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    CONSTRAINT ai_thread_scopes_status_check CHECK ((status = ANY (ARRAY['running'::text, 'waiting_approval'::text, 'succeeded'::text, 'failed'::text, 'canceled'::text])))
);


--
-- Name: TABLE ai_thread_scopes; Type: COMMENT; Schema: ai_agent_checkpoint; Owner: -
--

COMMENT ON TABLE ai_agent_checkpoint.ai_thread_scopes IS 'Safe checkpoint metadata used for namespace verification, retention and cleanup. Business payloads stay in LangGraph checkpoint/store tables.';


--
-- Name: action_attempts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.action_attempts (
    action_attempt_id uuid DEFAULT gen_random_uuid() NOT NULL,
    action_run_id uuid NOT NULL,
    attempt_no integer NOT NULL,
    status text NOT NULL,
    input_json jsonb,
    output_json jsonb,
    provider_response_json jsonb,
    error text,
    error_type text,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    finished_at timestamp with time zone,
    duration_ms integer,
    CONSTRAINT action_attempts_status_check CHECK ((status = ANY (ARRAY['running'::text, 'succeeded'::text, 'failed'::text])))
);


--
-- Name: action_dependencies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.action_dependencies (
    action_dependency_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid,
    action_dependency_type_id uuid NOT NULL,
    name text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    secret_id uuid,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    last_checked_at timestamp with time zone,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    company_id uuid,
    scope_kind text DEFAULT 'project'::text NOT NULL,
    CONSTRAINT action_dependencies_scope_check CHECK ((((scope_kind = 'company'::text) AND (company_id IS NOT NULL) AND (project_id IS NULL)) OR ((scope_kind = 'project'::text) AND (project_id IS NOT NULL)))),
    CONSTRAINT action_dependencies_scope_kind_check CHECK ((scope_kind = ANY (ARRAY['company'::text, 'project'::text]))),
    CONSTRAINT action_dependencies_status_check CHECK ((status = ANY (ARRAY['active'::text, 'disabled'::text, 'error'::text, 'deleted'::text])))
);


--
-- Name: action_dependency_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.action_dependency_types (
    action_dependency_type_id uuid DEFAULT gen_random_uuid() NOT NULL,
    dependency_type_key text NOT NULL,
    display_name text NOT NULL,
    provider_key text NOT NULL,
    auth_schema_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    metadata_schema_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: action_runs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.action_runs (
    action_run_id uuid DEFAULT gen_random_uuid() NOT NULL,
    workflow_run_id uuid NOT NULL,
    action_key text NOT NULL,
    action_type_key text NOT NULL,
    action_dependency_id uuid,
    status text DEFAULT 'pending'::text NOT NULL,
    attempt_count integer DEFAULT 0 NOT NULL,
    max_attempts integer,
    input_json jsonb,
    output_json jsonb,
    provider_response_json jsonb,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    next_retry_at timestamp with time zone,
    error text,
    retry_class text,
    retry_policy_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    retry_started_at timestamp with time zone,
    last_retry_at timestamp with time zone,
    retry_exhausted_at timestamp with time zone,
    dead_lettered_at timestamp with time zone,
    dead_letter_reason text,
    dead_letter_owner text,
    replayable boolean DEFAULT false NOT NULL,
    state_machine_version text DEFAULT 'action-run-state.v1'::text NOT NULL,
    terminal_reason text DEFAULT ''::text NOT NULL,
    cancelled_at timestamp with time zone,
    waiting_deadline_at timestamp with time zone,
    waiting_reason text DEFAULT ''::text NOT NULL,
    wake_reason text DEFAULT ''::text NOT NULL,
    last_state_transition_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT action_runs_retry_class_check CHECK (((retry_class IS NULL) OR (retry_class = ANY (ARRAY['validation'::text, 'auth'::text, 'permanent'::text, 'transient'::text, 'rate_limited'::text, 'timeout_unknown'::text, 'canceled'::text, 'unknown_outcome'::text])))),
    CONSTRAINT action_runs_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'running'::text, 'succeeded'::text, 'failed'::text, 'skipped'::text, 'waiting'::text, 'retrying'::text, 'dead_lettered'::text])))
);


--
-- Name: action_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.action_types (
    action_type_id uuid DEFAULT gen_random_uuid() NOT NULL,
    action_key text NOT NULL,
    display_name text NOT NULL,
    category text NOT NULL,
    config_schema_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    result_schema_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    requires_dependency boolean DEFAULT false NOT NULL,
    provider_key text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT action_types_category_check CHECK ((category = ANY (ARRAY['trigger'::text, 'logic'::text, 'control'::text, 'communication'::text, 'integration'::text])))
);


--
-- Name: ai_agent_audit_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_agent_audit_events (
    ai_agent_audit_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    user_id uuid,
    session_id uuid,
    run_id uuid,
    event_type text NOT NULL,
    risk_level text DEFAULT 'read'::text NOT NULL,
    tool_key text DEFAULT ''::text NOT NULL,
    decision text DEFAULT ''::text NOT NULL,
    input_hash text DEFAULT ''::text NOT NULL,
    output_hash text DEFAULT ''::text NOT NULL,
    redaction_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: ai_agent_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_agent_events (
    ai_agent_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    ai_agent_run_id uuid,
    ai_chat_session_id uuid NOT NULL,
    project_id uuid NOT NULL,
    seq bigint NOT NULL,
    event_type text NOT NULL,
    event_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: ai_agent_runs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_agent_runs (
    ai_agent_run_id uuid DEFAULT gen_random_uuid() NOT NULL,
    ai_chat_session_id uuid NOT NULL,
    project_id uuid NOT NULL,
    company_id uuid NOT NULL,
    user_id uuid,
    langgraph_thread_id text NOT NULL,
    runtime text DEFAULT 'langgraph'::text NOT NULL,
    status text NOT NULL,
    request_id text NOT NULL,
    idempotency_key text NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    finished_at timestamp with time zone,
    error_code text DEFAULT ''::text NOT NULL,
    error_text text DEFAULT ''::text NOT NULL,
    trace_id text DEFAULT ''::text NOT NULL,
    CONSTRAINT ai_agent_runs_status_check CHECK ((status = ANY (ARRAY['queued'::text, 'running'::text, 'waiting_approval'::text, 'succeeded'::text, 'failed'::text, 'canceled'::text])))
);


--
-- Name: ai_channel_bindings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_channel_bindings (
    ai_channel_binding_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_ai_employee_id uuid NOT NULL,
    project_id uuid NOT NULL,
    channel_key text NOT NULL,
    status text DEFAULT 'disabled'::text NOT NULL,
    action_dependency_id uuid,
    settings_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ai_channel_bindings_channel_check CHECK ((channel_key = ANY (ARRAY['web_widget'::text, 'telegram'::text, 'max'::text, 'site_chat'::text, 'email'::text, 'api'::text]))),
    CONSTRAINT ai_channel_bindings_status_check CHECK ((status = ANY (ARRAY['active'::text, 'disabled'::text, 'error'::text, 'deleted'::text])))
);


--
-- Name: ai_chat_cards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_chat_cards (
    ai_chat_card_id uuid DEFAULT gen_random_uuid() NOT NULL,
    ai_chat_session_id uuid NOT NULL,
    project_id uuid NOT NULL,
    card_id text NOT NULL,
    card_type text NOT NULL,
    revision integer DEFAULT 1 NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    risk_level text DEFAULT 'read'::text NOT NULL,
    title text DEFAULT ''::text NOT NULL,
    card_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ai_chat_cards_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'proposed'::text, 'needs_approval'::text, 'approved'::text, 'running'::text, 'succeeded'::text, 'failed'::text, 'rejected'::text, 'canceled'::text])))
);


--
-- Name: ai_chat_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_chat_messages (
    ai_chat_message_id uuid DEFAULT gen_random_uuid() NOT NULL,
    ai_chat_session_id uuid NOT NULL,
    role text NOT NULL,
    body text DEFAULT ''::text NOT NULL,
    payload_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    redaction_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    token_usage_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ai_chat_messages_role_check CHECK ((role = ANY (ARRAY['system'::text, 'user'::text, 'assistant'::text, 'tool'::text, 'approval'::text, 'event'::text])))
);


--
-- Name: ai_chat_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_chat_sessions (
    ai_chat_session_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_ai_employee_id uuid NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    screen_scope text DEFAULT ''::text NOT NULL,
    screen_entity_type text DEFAULT ''::text NOT NULL,
    screen_entity_id text DEFAULT ''::text NOT NULL,
    workflow_id uuid,
    workflow_run_id uuid,
    task_id uuid,
    client_id uuid,
    status text DEFAULT 'active'::text NOT NULL,
    action_dependency_id uuid,
    model_id text DEFAULT ''::text NOT NULL,
    created_by_user_id uuid,
    last_activity_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    page_context_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    last_page_context_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    ui_state_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    pending_context_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    thread_key text DEFAULT ''::text NOT NULL,
    active_tool_call_id uuid,
    active_patch_proposal_id uuid,
    session_revision integer DEFAULT 0 NOT NULL,
    active_intent_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    working_memory_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    langgraph_thread_id text DEFAULT ''::text NOT NULL,
    agent_runtime text DEFAULT 'go_legacy'::text NOT NULL,
    active_card_id text DEFAULT ''::text NOT NULL,
    CONSTRAINT ai_chat_sessions_status_check CHECK ((status = ANY (ARRAY['active'::text, 'running'::text, 'waiting_tool_approval'::text, 'waiting_patch_approval'::text, 'waiting_run_approval'::text, 'completed'::text, 'failed'::text, 'canceled'::text])))
);


--
-- Name: ai_model_catalog_cache; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_model_catalog_cache (
    ai_model_catalog_cache_id uuid DEFAULT gen_random_uuid() NOT NULL,
    action_dependency_id uuid NOT NULL,
    provider_key text NOT NULL,
    models_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    fetched_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone DEFAULT (now() + '00:15:00'::interval) NOT NULL,
    status text DEFAULT 'fresh'::text NOT NULL,
    error_text text DEFAULT ''::text NOT NULL
);


--
-- Name: ai_model_evaluations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_model_evaluations (
    ai_model_evaluation_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    action_dependency_id uuid NOT NULL,
    model_id text NOT NULL,
    provider_key text DEFAULT 'ai.apinet'::text NOT NULL,
    score numeric(8,4) DEFAULT 0 NOT NULL,
    capability_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    benchmark_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    selected_for text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: ai_patch_proposals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_patch_proposals (
    ai_patch_proposal_id uuid DEFAULT gen_random_uuid() NOT NULL,
    ai_chat_session_id uuid,
    project_id uuid NOT NULL,
    patch_target_type text NOT NULL,
    target_entity_id text DEFAULT ''::text NOT NULL,
    base_version_ref text DEFAULT ''::text NOT NULL,
    patch_format text NOT NULL,
    patch_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    summary text DEFAULT ''::text NOT NULL,
    warnings_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    status text DEFAULT 'proposed'::text NOT NULL,
    validated_at timestamp with time zone,
    applied_by_user_id uuid,
    applied_at timestamp with time zone,
    reverted_by_user_id uuid,
    reverted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ai_patch_proposals_status_check CHECK ((status = ANY (ARRAY['proposed'::text, 'validated'::text, 'pending_approval'::text, 'applied'::text, 'rejected'::text, 'failed'::text, 'reverted'::text])))
);


--
-- Name: ai_prompt_snapshots; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_prompt_snapshots (
    ai_prompt_snapshot_id uuid DEFAULT gen_random_uuid() NOT NULL,
    ai_chat_session_id uuid,
    ai_tool_call_id uuid,
    project_id uuid NOT NULL,
    provider_key text DEFAULT ''::text NOT NULL,
    model_id text DEFAULT ''::text NOT NULL,
    instructions text DEFAULT ''::text NOT NULL,
    prompt_text text DEFAULT ''::text NOT NULL,
    context_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    tools_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    redaction_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    compiled_at timestamp with time zone DEFAULT now() NOT NULL,
    copied_by_user_id uuid,
    copied_at timestamp with time zone
);


--
-- Name: ai_session_context_frames; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_session_context_frames (
    ai_session_context_frame_id uuid DEFAULT gen_random_uuid() NOT NULL,
    ai_chat_session_id uuid NOT NULL,
    project_id uuid NOT NULL,
    page_context_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    ui_state_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    client_context_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    active_tool_call_id uuid,
    active_patch_proposal_id uuid,
    local_date text DEFAULT ''::text NOT NULL,
    timezone text DEFAULT ''::text NOT NULL
);


--
-- Name: ai_tool_call_edits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_tool_call_edits (
    ai_tool_call_edit_id uuid DEFAULT gen_random_uuid() NOT NULL,
    ai_tool_call_id uuid NOT NULL,
    ai_chat_session_id uuid,
    project_id uuid NOT NULL,
    edited_by_user_id uuid,
    edit_reason text DEFAULT ''::text NOT NULL,
    before_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    after_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: ai_tool_calls; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_tool_calls (
    ai_tool_call_id uuid DEFAULT gen_random_uuid() NOT NULL,
    ai_chat_session_id uuid,
    project_id uuid NOT NULL,
    tool_key text NOT NULL,
    risk_level text DEFAULT 'read'::text NOT NULL,
    status text DEFAULT 'proposed'::text NOT NULL,
    input_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    output_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    error_text text DEFAULT ''::text NOT NULL,
    requires_approval boolean DEFAULT false NOT NULL,
    approved_by_user_id uuid,
    approved_at timestamp with time zone,
    rejected_by_user_id uuid,
    rejected_at timestamp with time zone,
    approval_reason text DEFAULT ''::text NOT NULL,
    idempotency_key text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    finished_at timestamp with time zone,
    CONSTRAINT ai_tool_calls_risk_level_check CHECK ((risk_level = ANY (ARRAY['read'::text, 'medium'::text, 'high'::text, 'critical'::text]))),
    CONSTRAINT ai_tool_calls_status_check CHECK ((status = ANY (ARRAY['proposed'::text, 'pending_approval'::text, 'approved'::text, 'rejected'::text, 'running'::text, 'succeeded'::text, 'failed'::text, 'canceled'::text])))
);


--
-- Name: ai_tool_idempotency; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_tool_idempotency (
    ai_tool_idempotency_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    idempotency_key text NOT NULL,
    tool_key text NOT NULL,
    input_hash text NOT NULL,
    status text NOT NULL,
    output_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    error_text text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    finished_at timestamp with time zone,
    CONSTRAINT ai_tool_idempotency_status_check CHECK ((status = ANY (ARRAY['reserved'::text, 'running'::text, 'succeeded'::text, 'failed'::text, 'conflict'::text])))
);


--
-- Name: ai_tool_plan_steps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_tool_plan_steps (
    ai_tool_plan_step_id uuid DEFAULT gen_random_uuid() NOT NULL,
    ai_tool_plan_id uuid NOT NULL,
    project_id uuid NOT NULL,
    step_id text NOT NULL,
    tool_key text NOT NULL,
    input_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    depends_on_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    output_refs_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    status text DEFAULT 'pending_approval'::text NOT NULL,
    idempotency_key text DEFAULT ''::text NOT NULL,
    output_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    error_text text DEFAULT ''::text NOT NULL,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: ai_tool_plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_tool_plans (
    ai_tool_plan_id uuid DEFAULT gen_random_uuid() NOT NULL,
    ai_chat_session_id uuid NOT NULL,
    project_id uuid NOT NULL,
    company_id uuid NOT NULL,
    user_id uuid,
    plan_key text DEFAULT ''::text NOT NULL,
    intent_key text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    plan_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    idempotency_key text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: ai_usage_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_usage_events (
    ai_usage_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    ai_chat_session_id uuid,
    workflow_run_id uuid,
    action_run_id uuid,
    provider_key text DEFAULT ''::text NOT NULL,
    model_id text DEFAULT ''::text NOT NULL,
    input_tokens integer DEFAULT 0 NOT NULL,
    output_tokens integer DEFAULT 0 NOT NULL,
    cached_tokens integer DEFAULT 0 NOT NULL,
    reasoning_tokens integer DEFAULT 0 NOT NULL,
    cost_estimate numeric,
    latency_ms integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    provider_response_json jsonb DEFAULT '{}'::jsonb NOT NULL
);


--
-- Name: api_key_grants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.api_key_grants (
    api_key_grant_id uuid DEFAULT gen_random_uuid() NOT NULL,
    api_key_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid,
    permissions text[] DEFAULT ARRAY[]::text[] NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT api_key_grants_permissions_check CHECK ((permissions <@ ARRAY['read'::text, 'write'::text, 'admin'::text])),
    CONSTRAINT api_key_grants_scope_id_check CHECK ((((scope_kind = 'system'::text) AND (scope_id IS NULL)) OR ((scope_kind <> 'system'::text) AND (scope_id IS NOT NULL)))),
    CONSTRAINT api_key_grants_scope_kind_check CHECK ((scope_kind = ANY (ARRAY['system'::text, 'company'::text, 'project'::text, 'workflow'::text, 'client'::text, 'client_service'::text, 'payment_action'::text])))
);


--
-- Name: api_keys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.api_keys (
    api_key_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    name text NOT NULL,
    key_prefix text NOT NULL,
    key_hash text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    last_used_at timestamp with time zone,
    expires_at timestamp with time zone,
    revoked_at timestamp with time zone,
    revoked_reason text,
    CONSTRAINT api_keys_status_check CHECK ((status = ANY (ARRAY['active'::text, 'revoked'::text])))
);


--
-- Name: audit_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_events (
    audit_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid,
    actor_type text NOT NULL,
    actor_id text,
    event_type text NOT NULL,
    payload_json jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: auth_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.auth_events (
    auth_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    event_type text NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: auth_identities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.auth_identities (
    auth_identity_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    provider text NOT NULL,
    provider_subject text NOT NULL,
    provider_email public.citext,
    provider_login text,
    is_primary boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT auth_identities_provider_check CHECK ((provider = ANY (ARRAY['password'::text, 'telegram'::text, 'google'::text, 'yandex'::text, 'vk'::text])))
);


--
-- Name: auth_identity_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.auth_identity_versions (
    auth_identity_version_id uuid DEFAULT gen_random_uuid() NOT NULL,
    auth_identity_id uuid NOT NULL,
    version_no integer NOT NULL,
    operation_type text NOT NULL,
    snapshot jsonb NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by_user_id uuid,
    request_id text,
    CONSTRAINT auth_identity_versions_operation_type_check CHECK ((operation_type = ANY (ARRAY['insert'::text, 'update'::text, 'delete'::text])))
);


--
-- Name: automation_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.automation_rules (
    automation_rule_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid,
    name text NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    trigger text NOT NULL,
    conditions_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    actions_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    scope text DEFAULT 'project'::text NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT automation_rules_scope_check CHECK ((scope = ANY (ARRAY['workspace'::text, 'project'::text, 'client'::text, 'task_type'::text])))
);


--
-- Name: bot_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.bot_sessions (
    bot_session_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    action_dependency_id uuid NOT NULL,
    relay_key text NOT NULL,
    chat_id text NOT NULL,
    user_id text DEFAULT ''::text NOT NULL,
    state_key text DEFAULT ''::text NOT NULL,
    context_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    last_update_id bigint,
    last_telegram_relay_event_id uuid,
    last_workflow_run_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: client_contacts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.client_contacts (
    client_contact_id uuid DEFAULT gen_random_uuid() NOT NULL,
    client_id uuid NOT NULL,
    company_id uuid NOT NULL,
    first_name text,
    last_name text,
    middle_name text,
    "position" text,
    email text,
    phone text,
    telegram text,
    preferred_channel text DEFAULT 'email'::text NOT NULL,
    comment text,
    is_primary boolean DEFAULT false NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT client_contacts_channel_check CHECK ((preferred_channel = ANY (ARRAY['email'::text, 'phone'::text, 'telegram'::text, 'any'::text]))),
    CONSTRAINT client_contacts_status_check CHECK ((status = ANY (ARRAY['active'::text, 'archived'::text, 'deleted'::text])))
);


--
-- Name: client_projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.client_projects (
    client_id uuid NOT NULL,
    project_id uuid NOT NULL,
    company_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: client_responsibles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.client_responsibles (
    client_responsible_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    client_id uuid NOT NULL,
    user_id uuid,
    company_legal_entity_id uuid,
    display_name text NOT NULL,
    role text DEFAULT 'account_manager'::text NOT NULL,
    is_primary boolean DEFAULT false NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    comment text,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT client_responsibles_role_check CHECK ((role = ANY (ARRAY['account_manager'::text, 'technical_lead'::text, 'finance'::text, 'legal'::text, 'owner'::text, 'other'::text]))),
    CONSTRAINT client_responsibles_status_check CHECK ((status = ANY (ARRAY['active'::text, 'archived'::text, 'deleted'::text])))
);


--
-- Name: client_service_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.client_service_types (
    client_service_type_id uuid DEFAULT gen_random_uuid() NOT NULL,
    service_type_key text NOT NULL,
    display_name text NOT NULL,
    description text,
    status text DEFAULT 'active'::text NOT NULL,
    sort_order integer DEFAULT 100 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    company_id uuid,
    project_id uuid,
    scope_kind text DEFAULT 'global'::text NOT NULL,
    CONSTRAINT client_service_types_scope_kind_check CHECK ((scope_kind = ANY (ARRAY['global'::text, 'project'::text]))),
    CONSTRAINT client_service_types_status_check CHECK ((status = ANY (ARRAY['active'::text, 'archived'::text])))
);


--
-- Name: client_services; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.client_services (
    client_service_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    client_id uuid NOT NULL,
    project_id uuid NOT NULL,
    client_service_type_id uuid NOT NULL,
    name text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    amount numeric(14,2),
    raw_price_text text,
    currency text DEFAULT 'RUB'::text NOT NULL,
    billing_period text DEFAULT 'monthly'::text NOT NULL,
    next_due_date date,
    remind_days_before integer DEFAULT 7 NOT NULL,
    remind_every_days integer DEFAULT 7 NOT NULL,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT client_services_billing_period_check CHECK ((billing_period = ANY (ARRAY['once'::text, 'weekly'::text, 'monthly'::text, 'quarterly'::text, 'semiannual'::text, 'yearly'::text]))),
    CONSTRAINT client_services_currency_check CHECK ((currency = 'RUB'::text)),
    CONSTRAINT client_services_reminders_check CHECK (((remind_days_before >= 0) AND (remind_days_before <= 90) AND (remind_every_days >= 1) AND (remind_every_days <= 365))),
    CONSTRAINT client_services_status_check CHECK ((status = ANY (ARRAY['active'::text, 'paused'::text, 'completed'::text, 'archived'::text, 'deleted'::text])))
);


--
-- Name: clients; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.clients (
    client_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    client_type text DEFAULT 'legal_entity'::text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    name text NOT NULL,
    short_name text,
    full_name text,
    inn text,
    kpp text,
    ogrn text,
    ogrnip text,
    legal_address text,
    actual_address text,
    okved text,
    dadata_raw_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    comment text,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    legal_name text,
    industry text,
    website text,
    email text,
    phone text,
    telegram text,
    manager_user_id uuid,
    account_manager_user_id uuid,
    source text,
    tags text[] DEFAULT ARRAY[]::text[] NOT NULL,
    search_vector tsvector,
    CONSTRAINT clients_status_check CHECK ((status = ANY (ARRAY['lead'::text, 'active'::text, 'paused'::text, 'archived'::text, 'deleted'::text]))),
    CONSTRAINT clients_type_check CHECK ((client_type = ANY (ARRAY['individual'::text, 'individual_entrepreneur'::text, 'legal_entity'::text])))
);


--
-- Name: companies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.companies (
    company_id uuid DEFAULT gen_random_uuid() NOT NULL,
    owner_user_id uuid NOT NULL,
    slug text,
    name text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by_user_id uuid,
    logo_url text,
    timezone text DEFAULT 'Europe/Moscow'::text NOT NULL,
    default_currency text DEFAULT 'RUB'::text NOT NULL,
    workday_start time without time zone DEFAULT '10:00:00'::time without time zone NOT NULL,
    workday_end time without time zone DEFAULT '19:00:00'::time without time zone NOT NULL,
    workweek_days integer[] DEFAULT ARRAY[1, 2, 3, 4, 5] NOT NULL,
    CONSTRAINT companies_status_check CHECK ((status = ANY (ARRAY['active'::text, 'suspended'::text, 'archived'::text])))
);


--
-- Name: company_legal_entities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.company_legal_entities (
    company_legal_entity_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    entity_type text DEFAULT 'legal_entity'::text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    name text NOT NULL,
    short_name text,
    full_name text,
    inn text,
    kpp text,
    ogrn text,
    ogrnip text,
    legal_address text,
    actual_address text,
    okved text,
    dadata_raw_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    bank_settings_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    comment text,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT company_legal_entities_status_check CHECK ((status = ANY (ARRAY['active'::text, 'archived'::text, 'deleted'::text]))),
    CONSTRAINT company_legal_entities_type_check CHECK ((entity_type = ANY (ARRAY['legal_entity'::text, 'individual_entrepreneur'::text])))
);


--
-- Name: company_membership_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.company_membership_versions (
    company_membership_version_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    user_id uuid NOT NULL,
    version_no integer NOT NULL,
    operation_type text NOT NULL,
    snapshot jsonb NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by_user_id uuid,
    request_id text,
    CONSTRAINT company_membership_versions_operation_type_check CHECK ((operation_type = ANY (ARRAY['insert'::text, 'update'::text, 'delete'::text])))
);


--
-- Name: company_memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.company_memberships (
    company_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role text DEFAULT 'member'::text NOT NULL,
    permissions jsonb DEFAULT '{}'::jsonb NOT NULL,
    invited_by uuid,
    invited_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT company_memberships_role_check CHECK ((role = ANY (ARRAY['owner'::text, 'admin'::text, 'member'::text])))
);


--
-- Name: company_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.company_versions (
    company_version_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    version_no integer NOT NULL,
    operation_type text NOT NULL,
    snapshot jsonb NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by_user_id uuid,
    request_id text,
    CONSTRAINT company_versions_operation_type_check CHECK ((operation_type = ANY (ARRAY['insert'::text, 'update'::text, 'delete'::text])))
);


--
-- Name: desktop_auth_requests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.desktop_auth_requests (
    desktop_auth_request_id uuid DEFAULT gen_random_uuid() NOT NULL,
    device_code text NOT NULL,
    user_code text NOT NULL,
    device_name text,
    status text DEFAULT 'pending'::text NOT NULL,
    user_id uuid,
    project_id uuid,
    desktop_device_id uuid,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    approved_at timestamp with time zone,
    CONSTRAINT desktop_auth_requests_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'approved'::text, 'denied'::text, 'expired'::text])))
);


--
-- Name: desktop_devices; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.desktop_devices (
    desktop_device_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    project_id uuid NOT NULL,
    name text NOT NULL,
    device_key text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    last_seen_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    paired_at timestamp with time zone,
    revoked_at timestamp with time zone,
    CONSTRAINT desktop_devices_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'active'::text, 'revoked'::text])))
);


--
-- Name: desktop_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.desktop_messages (
    desktop_message_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    desktop_device_id uuid,
    message_type text NOT NULL,
    title text NOT NULL,
    body text,
    target_url text,
    target_workflow_id uuid,
    target_workflow_run_id uuid,
    target_webhook_event_id uuid,
    status text DEFAULT 'new'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    delivered_at timestamp with time zone,
    acknowledged_at timestamp with time zone,
    CONSTRAINT desktop_messages_message_type_check CHECK ((message_type = ANY (ARRAY['open_workflow'::text, 'open_run'::text, 'open_event'::text, 'info'::text]))),
    CONSTRAINT desktop_messages_status_check CHECK ((status = ANY (ARRAY['new'::text, 'delivered'::text, 'acknowledged'::text])))
);


--
-- Name: desktop_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.desktop_tokens (
    desktop_token_id uuid DEFAULT gen_random_uuid() NOT NULL,
    desktop_device_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    revoked_at timestamp with time zone
);


--
-- Name: device_credential_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.device_credential_versions (
    device_credential_version_id uuid DEFAULT gen_random_uuid() NOT NULL,
    device_credential_id uuid NOT NULL,
    version_no integer NOT NULL,
    operation_type text NOT NULL,
    snapshot jsonb NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by_user_id uuid,
    request_id text,
    CONSTRAINT device_credential_versions_operation_type_check CHECK ((operation_type = ANY (ARRAY['insert'::text, 'update'::text, 'delete'::text])))
);


--
-- Name: device_credentials; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.device_credentials (
    device_credential_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    session_id uuid,
    device_platform text NOT NULL,
    device_fingerprint_hash text NOT NULL,
    credential_hash text NOT NULL,
    issued_at timestamp with time zone DEFAULT now() NOT NULL,
    last_used_at timestamp with time zone,
    rotate_after timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT device_credentials_expires_after_issued_check CHECK ((expires_at > issued_at)),
    CONSTRAINT device_credentials_platform_check CHECK ((device_platform = ANY (ARRAY['qt-windows'::text, 'qt-macos'::text, 'qt-linux'::text, 'android-webview'::text]))),
    CONSTRAINT device_credentials_rotate_before_expires_check CHECK ((rotate_after < expires_at))
);


--
-- Name: effect_ledger; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.effect_ledger (
    effect_id text NOT NULL,
    workflow_run_id uuid NOT NULL,
    action_run_id uuid,
    action_key text NOT NULL,
    action_type_key text NOT NULL,
    attempt_no integer NOT NULL,
    idempotency_key text NOT NULL,
    provider_idempotency_key text DEFAULT ''::text NOT NULL,
    input_fingerprint text DEFAULT ''::text NOT NULL,
    status text NOT NULL,
    non_idempotent boolean DEFAULT true NOT NULL,
    output_json jsonb,
    provider_response_json jsonb,
    response_fingerprint text DEFAULT ''::text NOT NULL,
    error_type text DEFAULT ''::text NOT NULL,
    error text DEFAULT ''::text NOT NULL,
    reserved_at timestamp with time zone DEFAULT now() NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    finished_at timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT effect_ledger_attempt_no_check CHECK ((attempt_no > 0)),
    CONSTRAINT effect_ledger_status_check CHECK ((status = ANY (ARRAY['started'::text, 'succeeded'::text, 'failed'::text, 'unknown'::text])))
);


--
-- Name: email_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_templates (
    email_template_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    subject_template text NOT NULL,
    body_template text NOT NULL,
    body_format text DEFAULT 'html'::text NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT email_templates_body_format_check CHECK ((body_format = ANY (ARRAY['text'::text, 'html'::text, 'markdown'::text])))
);


--
-- Name: employee_profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.employee_profiles (
    employee_profile_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    company_id uuid NOT NULL,
    full_name text NOT NULL,
    avatar_url text,
    email text,
    phone text,
    telegram text,
    "position" text,
    department text,
    role text DEFAULT 'employee'::text NOT NULL,
    employment_type text DEFAULT 'full_time'::text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    timezone text DEFAULT 'Europe/Moscow'::text NOT NULL,
    capacity_hours_per_week integer DEFAULT 40 NOT NULL,
    cost_rate numeric(12,2),
    bill_rate numeric(12,2),
    skills text[] DEFAULT ARRAY[]::text[] NOT NULL,
    seniority text,
    manager_employee_profile_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT employee_profiles_capacity_check CHECK (((capacity_hours_per_week >= 0) AND (capacity_hours_per_week <= 80))),
    CONSTRAINT employee_profiles_employment_type_check CHECK ((employment_type = ANY (ARRAY['full_time'::text, 'part_time'::text, 'contractor'::text, 'client_guest'::text]))),
    CONSTRAINT employee_profiles_role_check CHECK ((role = ANY (ARRAY['owner'::text, 'admin'::text, 'operations_manager'::text, 'project_manager'::text, 'account_manager'::text, 'team_lead'::text, 'employee'::text, 'contractor'::text, 'client_guest'::text, 'finance'::text, 'hr'::text, 'read_only_auditor'::text]))),
    CONSTRAINT employee_profiles_status_check CHECK ((status = ANY (ARRAY['active'::text, 'inactive'::text, 'vacation'::text, 'fired'::text])))
);


--
-- Name: integration_doc_chunks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.integration_doc_chunks (
    integration_doc_chunk_id uuid DEFAULT gen_random_uuid() NOT NULL,
    integration_doc_id uuid NOT NULL,
    chunk_index integer NOT NULL,
    heading text DEFAULT ''::text NOT NULL,
    content text NOT NULL,
    search_vector tsvector DEFAULT ''::tsvector NOT NULL,
    embedding_model text,
    embedding_json jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: integration_docs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.integration_docs (
    integration_doc_id uuid DEFAULT gen_random_uuid() NOT NULL,
    doc_key text NOT NULL,
    public_slug text NOT NULL,
    title text NOT NULL,
    summary text NOT NULL,
    provider_key text DEFAULT 'internal'::text NOT NULL,
    action_type_keys text[] DEFAULT '{}'::text[] NOT NULL,
    dependency_type_keys text[] DEFAULT '{}'::text[] NOT NULL,
    tags text[] DEFAULT '{}'::text[] NOT NULL,
    intents text[] DEFAULT '{}'::text[] NOT NULL,
    markdown text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    embedding_model text,
    embedding_json jsonb,
    search_vector tsvector DEFAULT ''::tsvector NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT integration_docs_status_check CHECK ((status = ANY (ARRAY['active'::text, 'draft'::text, 'archived'::text])))
);


--
-- Name: message_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.message_templates (
    message_template_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    provider_key text NOT NULL,
    body_template text NOT NULL,
    body_format text DEFAULT 'markdown'::text NOT NULL,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT message_templates_body_format_check CHECK ((body_format = ANY (ARRAY['text'::text, 'html'::text, 'markdown'::text])))
);


--
-- Name: module_channel_bindings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.module_channel_bindings (
    module_channel_binding_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid,
    module_key text NOT NULL,
    event_key text NOT NULL,
    channel_key text NOT NULL,
    action_type_key text NOT NULL,
    action_dependency_id uuid,
    message_template_id uuid,
    email_template_id uuid,
    enabled boolean DEFAULT false NOT NULL,
    config_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT module_channel_bindings_channel_key_check CHECK ((btrim(channel_key) <> ''::text)),
    CONSTRAINT module_channel_bindings_event_key_check CHECK ((btrim(event_key) <> ''::text)),
    CONSTRAINT module_channel_bindings_module_key_check CHECK ((btrim(module_key) <> ''::text))
);


--
-- Name: module_workflow_triggers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.module_workflow_triggers (
    module_workflow_trigger_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid,
    module_key text NOT NULL,
    event_key text NOT NULL,
    workflow_id uuid NOT NULL,
    workflow_version_id uuid,
    enabled boolean DEFAULT true NOT NULL,
    config_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT module_workflow_triggers_event_key_check CHECK ((btrim(event_key) <> ''::text)),
    CONSTRAINT module_workflow_triggers_module_key_check CHECK ((btrim(module_key) <> ''::text))
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    notification_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    user_id uuid NOT NULL,
    task_id uuid,
    event_type text NOT NULL,
    title text NOT NULL,
    body text,
    read_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: outbox_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.outbox_events (
    outbox_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    event_type text NOT NULL,
    aggregate_type text DEFAULT ''::text NOT NULL,
    aggregate_id text DEFAULT ''::text NOT NULL,
    idempotency_key text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    payload_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    max_attempts integer DEFAULT 10 NOT NULL,
    available_at timestamp with time zone DEFAULT now() NOT NULL,
    locked_until timestamp with time zone,
    last_error text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT outbox_events_attempts_check CHECK ((attempts >= 0)),
    CONSTRAINT outbox_events_max_attempts_check CHECK ((max_attempts > 0)),
    CONSTRAINT outbox_events_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'processing'::text, 'processed'::text, 'failed'::text])))
);


--
-- Name: payment_action_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payment_action_documents (
    payment_action_document_id uuid DEFAULT gen_random_uuid() NOT NULL,
    payment_action_id uuid NOT NULL,
    company_id uuid NOT NULL,
    client_id uuid NOT NULL,
    project_id uuid NOT NULL,
    uploaded_by_user_id uuid,
    file_name text NOT NULL,
    content_type text DEFAULT 'application/octet-stream'::text NOT NULL,
    size_bytes bigint DEFAULT 0 NOT NULL,
    kind text DEFAULT 'file'::text NOT NULL,
    storage_path text NOT NULL,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT payment_action_documents_kind_check CHECK ((kind = ANY (ARRAY['invoice'::text, 'act'::text, 'receipt'::text, 'reconciliation'::text, 'file'::text])))
);


--
-- Name: payment_action_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payment_action_items (
    payment_action_item_id uuid DEFAULT gen_random_uuid() NOT NULL,
    payment_action_id uuid NOT NULL,
    client_service_id uuid NOT NULL,
    client_service_type_id uuid,
    title text NOT NULL,
    due_date date NOT NULL,
    amount numeric(14,2) DEFAULT 0 NOT NULL,
    currency text DEFAULT 'RUB'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT payment_action_items_currency_check CHECK ((currency = 'RUB'::text))
);


--
-- Name: payment_actions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payment_actions (
    payment_action_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    client_id uuid NOT NULL,
    project_id uuid NOT NULL,
    workflow_run_id uuid,
    status text DEFAULT 'pending'::text NOT NULL,
    source text DEFAULT 'manual'::text NOT NULL,
    period_start date NOT NULL,
    period_end date NOT NULL,
    due_date date NOT NULL,
    total_amount numeric(14,2) DEFAULT 0 NOT NULL,
    currency text DEFAULT 'RUB'::text NOT NULL,
    last_reminded_at timestamp with time zone,
    remind_after_at timestamp with time zone,
    snoozed_until timestamp with time zone,
    confirmed_at timestamp with time zone,
    comment text,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    payment_method text DEFAULT 'legal_invoice'::text NOT NULL,
    CONSTRAINT payment_actions_currency_check CHECK ((currency = 'RUB'::text)),
    CONSTRAINT payment_actions_method_check CHECK ((payment_method = ANY (ARRAY['cash'::text, 'legal_invoice'::text, 'bank_transfer'::text, 'card'::text, 'other'::text]))),
    CONSTRAINT payment_actions_source_check CHECK ((source = ANY (ARRAY['manual'::text, 'api_key'::text, 'bank'::text, 'workflow'::text, 'email'::text, 'incoming_email'::text, 'kontur_elba'::text, 'integration'::text, 'other_integration'::text]))),
    CONSTRAINT payment_actions_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'sent'::text, 'snoozed'::text, 'paid'::text, 'cancelled'::text])))
);


--
-- Name: payment_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payment_events (
    payment_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    event_type text NOT NULL,
    event_source text NOT NULL,
    event_subject text NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL,
    idempotency_key text NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    client_id uuid NOT NULL,
    client_service_id uuid,
    payment_action_id uuid,
    payment_fact_id uuid,
    workflow_run_id uuid,
    amount numeric(14,2),
    currency text DEFAULT 'RUB'::text NOT NULL,
    period_start date,
    period_end date,
    paid_at timestamp with time zone,
    source text NOT NULL,
    trigger_type text NOT NULL,
    status text DEFAULT 'processing'::text NOT NULL,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    error text,
    created_by_user_id uuid,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    payment_method text DEFAULT 'legal_invoice'::text NOT NULL,
    CONSTRAINT payment_events_currency_check CHECK ((currency = 'RUB'::text)),
    CONSTRAINT payment_events_idempotency_key_check CHECK ((length(btrim(idempotency_key)) > 0)),
    CONSTRAINT payment_events_method_check CHECK ((payment_method = ANY (ARRAY['cash'::text, 'legal_invoice'::text, 'bank_transfer'::text, 'card'::text, 'other'::text]))),
    CONSTRAINT payment_events_source_check CHECK ((source = ANY (ARRAY['manual'::text, 'api_key'::text, 'bank'::text, 'workflow'::text, 'email'::text, 'incoming_email'::text, 'kontur_elba'::text, 'integration'::text, 'other_integration'::text]))),
    CONSTRAINT payment_events_status_check CHECK ((status = ANY (ARRAY['processing'::text, 'processed'::text, 'failed'::text])))
);


--
-- Name: payment_facts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payment_facts (
    payment_fact_id uuid DEFAULT gen_random_uuid() NOT NULL,
    payment_action_id uuid NOT NULL,
    company_id uuid NOT NULL,
    client_id uuid NOT NULL,
    project_id uuid NOT NULL,
    source text DEFAULT 'manual'::text NOT NULL,
    amount numeric(14,2) NOT NULL,
    currency text DEFAULT 'RUB'::text NOT NULL,
    paid_at timestamp with time zone DEFAULT now() NOT NULL,
    raw_payload_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    payment_method text DEFAULT 'legal_invoice'::text NOT NULL,
    CONSTRAINT payment_facts_currency_check CHECK ((currency = 'RUB'::text)),
    CONSTRAINT payment_facts_method_check CHECK ((payment_method = ANY (ARRAY['cash'::text, 'legal_invoice'::text, 'bank_transfer'::text, 'card'::text, 'other'::text]))),
    CONSTRAINT payment_facts_source_check CHECK ((source = ANY (ARRAY['manual'::text, 'api_key'::text, 'bank'::text, 'workflow'::text, 'email'::text, 'incoming_email'::text, 'kontur_elba'::text, 'integration'::text, 'other_integration'::text])))
);


--
-- Name: project_ai_employees; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_ai_employees (
    project_ai_employee_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    company_id uuid NOT NULL,
    employee_key text DEFAULT 'default'::text NOT NULL,
    display_name text DEFAULT 'AI сотрудник'::text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    action_dependency_id uuid,
    model_id text DEFAULT ''::text NOT NULL,
    profile_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    tool_policy_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    memory_policy_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    ui_settings_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT project_ai_employees_status_check CHECK ((status = ANY (ARRAY['active'::text, 'disabled'::text, 'archived'::text])))
);


--
-- Name: project_member_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_member_versions (
    project_member_version_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    user_id uuid NOT NULL,
    version_no integer NOT NULL,
    operation_type text NOT NULL,
    snapshot jsonb NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by_user_id uuid,
    request_id text,
    CONSTRAINT project_member_versions_operation_type_check CHECK ((operation_type = ANY (ARRAY['insert'::text, 'update'::text, 'delete'::text])))
);


--
-- Name: project_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_members (
    project_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role text DEFAULT 'owner'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    permissions jsonb DEFAULT '{}'::jsonb NOT NULL,
    client_id uuid,
    invited_by_user_id uuid,
    invited_at timestamp with time zone,
    status text DEFAULT 'active'::text NOT NULL,
    CONSTRAINT project_members_role_check CHECK ((role = ANY (ARRAY['owner'::text, 'admin'::text, 'operations_manager'::text, 'project_manager'::text, 'account_manager'::text, 'team_lead'::text, 'employee'::text, 'contractor'::text, 'client_guest'::text, 'finance'::text, 'hr'::text, 'read_only_auditor'::text, 'member'::text]))),
    CONSTRAINT project_members_status_check CHECK ((status = ANY (ARRAY['invited'::text, 'active'::text, 'inactive'::text, 'removed'::text])))
);


--
-- Name: project_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_versions (
    project_version_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    version_no integer NOT NULL,
    operation_type text NOT NULL,
    snapshot jsonb NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by_user_id uuid,
    request_id text,
    CONSTRAINT project_versions_operation_type_check CHECK ((operation_type = ANY (ARRAY['insert'::text, 'update'::text, 'delete'::text])))
);


--
-- Name: projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.projects (
    project_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    hook_key text NOT NULL,
    hook_secret_hash text NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    project_key text,
    description text,
    status text DEFAULT 'active'::text NOT NULL,
    health text DEFAULT 'green'::text NOT NULL,
    priority text DEFAULT 'medium'::text NOT NULL,
    start_date date,
    target_date date,
    deadline date,
    budget_hours numeric(10,2),
    budget_money numeric(14,2),
    billing_type text DEFAULT 'non_billable'::text NOT NULL,
    owner_user_id uuid,
    project_manager_user_id uuid,
    task_workflow_id uuid,
    CONSTRAINT projects_billing_type_check CHECK ((billing_type = ANY (ARRAY['non_billable'::text, 'hourly'::text, 'fixed'::text, 'retainer'::text]))),
    CONSTRAINT projects_health_check CHECK ((health = ANY (ARRAY['green'::text, 'yellow'::text, 'red'::text]))),
    CONSTRAINT projects_priority_check CHECK ((priority = ANY (ARRAY['none'::text, 'low'::text, 'medium'::text, 'high'::text, 'urgent'::text]))),
    CONSTRAINT projects_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'active'::text, 'on_hold'::text, 'completed'::text, 'archived'::text])))
);


--
-- Name: provider_operation_effects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.provider_operation_effects (
    provider_operation_effect_id uuid DEFAULT gen_random_uuid() NOT NULL,
    action_dependency_id uuid NOT NULL,
    provider text NOT NULL,
    operation_key text NOT NULL,
    request_sha256 text NOT NULL,
    status text NOT NULL,
    request_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    result_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    possible_effect boolean DEFAULT false NOT NULL,
    retryable boolean DEFAULT false NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    last_error_code text DEFAULT ''::text NOT NULL,
    finished_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT provider_operation_effects_attempts_check CHECK ((attempts >= 0)),
    CONSTRAINT provider_operation_effects_provider_check CHECK ((provider = 'yclients'::text)),
    CONSTRAINT provider_operation_effects_request_sha_check CHECK ((request_sha256 ~ '^[0-9a-f]{64}$'::text)),
    CONSTRAINT provider_operation_effects_status_check CHECK ((status = ANY (ARRAY['planned'::text, 'sending'::text, 'succeeded'::text, 'failed'::text, 'unknown'::text])))
);


--
-- Name: run_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.run_events (
    run_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    workflow_run_id uuid NOT NULL,
    action_key text,
    event_type text NOT NULL,
    message text,
    payload_json jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: secrets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.secrets (
    secret_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid,
    name text NOT NULL,
    encrypted_value bytea NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    company_id uuid,
    scope_kind text DEFAULT 'project'::text NOT NULL,
    CONSTRAINT secrets_scope_check CHECK ((((scope_kind = 'company'::text) AND (company_id IS NOT NULL) AND (project_id IS NULL)) OR ((scope_kind = 'project'::text) AND (project_id IS NOT NULL)))),
    CONSTRAINT secrets_scope_kind_check CHECK ((scope_kind = ANY (ARRAY['company'::text, 'project'::text])))
);


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sessions (
    session_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    session_token_hash text NOT NULL,
    device_label text,
    user_agent text,
    ip_address inet,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_seen_at timestamp with time zone,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    refresh_token_hash text NOT NULL,
    client_type text DEFAULT 'web'::text NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    revoked_reason text,
    access_expires_at timestamp with time zone,
    CONSTRAINT sessions_client_type_check CHECK ((client_type = ANY (ARRAY['web'::text, 'telegram-web'::text, 'qt-windows'::text, 'qt-macos'::text, 'qt-linux'::text, 'android-webview'::text]))),
    CONSTRAINT sessions_expires_after_created_check CHECK ((expires_at > created_at))
);


--
-- Name: site_call_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.site_call_sessions (
    call_session_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    chat_widget_id uuid NOT NULL,
    visitor_id uuid NOT NULL,
    visit_session_id uuid,
    conversation_id uuid,
    direction text NOT NULL,
    requested_by text NOT NULL,
    requested_by_user_id uuid,
    status text DEFAULT 'offered'::text NOT NULL,
    privacy_consent_accepted_at timestamp with time zone,
    recording_consent_accepted_at timestamp with time zone,
    recording_enabled boolean DEFAULT false NOT NULL,
    media_kind text DEFAULT 'audio'::text NOT NULL,
    started_at timestamp with time zone,
    ended_at timestamp with time zone,
    last_signal_at timestamp with time zone,
    end_reason text DEFAULT ''::text NOT NULL,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT site_call_sessions_direction_check CHECK ((direction = ANY (ARRAY['operator_to_visitor'::text, 'visitor_to_operator'::text]))),
    CONSTRAINT site_call_sessions_media_kind_check CHECK ((media_kind = ANY (ARRAY['audio'::text, 'video'::text]))),
    CONSTRAINT site_call_sessions_requested_by_check CHECK ((requested_by = ANY (ARRAY['operator'::text, 'visitor'::text]))),
    CONSTRAINT site_call_sessions_status_check CHECK ((status = ANY (ARRAY['offered'::text, 'ringing'::text, 'accepted'::text, 'rejected'::text, 'missed'::text, 'ended'::text, 'failed'::text, 'cancelled'::text])))
);


--
-- Name: site_chat_attachments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.site_chat_attachments (
    attachment_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    chat_widget_id uuid,
    conversation_id uuid NOT NULL,
    message_id uuid,
    call_session_id uuid,
    uploaded_by_kind text NOT NULL,
    uploaded_by_user_id uuid,
    file_name text NOT NULL,
    content_type text DEFAULT 'application/octet-stream'::text NOT NULL,
    size_bytes bigint NOT NULL,
    storage_path text NOT NULL,
    download_token text NOT NULL,
    kind text NOT NULL,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT site_chat_attachments_kind_check CHECK ((kind = ANY (ARRAY['image'::text, 'audio'::text, 'video'::text, 'file'::text]))),
    CONSTRAINT site_chat_attachments_size_bytes_check CHECK ((size_bytes >= 0)),
    CONSTRAINT site_chat_attachments_uploaded_by_kind_check CHECK ((uploaded_by_kind = ANY (ARRAY['visitor'::text, 'operator'::text, 'system'::text])))
);


--
-- Name: site_chat_conversations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.site_chat_conversations (
    conversation_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    chat_widget_id uuid NOT NULL,
    visitor_id uuid NOT NULL,
    visit_session_id uuid,
    status text DEFAULT 'open'::text NOT NULL,
    title text DEFAULT 'Чат с посетителем'::text NOT NULL,
    visitor_name text DEFAULT ''::text NOT NULL,
    visitor_contact text DEFAULT ''::text NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT site_chat_conversations_status_check CHECK ((status = ANY (ARRAY['open'::text, 'closed'::text])))
);


--
-- Name: site_chat_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.site_chat_messages (
    message_id uuid DEFAULT gen_random_uuid() NOT NULL,
    conversation_id uuid NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    author_kind text NOT NULL,
    author_user_id uuid,
    author_name text DEFAULT ''::text NOT NULL,
    body text NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT site_chat_messages_author_kind_check CHECK ((author_kind = ANY (ARRAY['visitor'::text, 'operator'::text, 'system'::text]))),
    CONSTRAINT site_chat_messages_body_check CHECK ((btrim(body) <> ''::text))
);


--
-- Name: site_chat_widgets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.site_chat_widgets (
    chat_widget_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    public_key text NOT NULL,
    name text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    title text DEFAULT 'Чат на сайте'::text NOT NULL,
    subtitle text DEFAULT 'Напишите нам, оператор ответит здесь.'::text NOT NULL,
    accent_color text DEFAULT '#00B2FF'::text NOT NULL,
    "position" text DEFAULT 'right'::text NOT NULL,
    heavy_tracking_default boolean DEFAULT false NOT NULL,
    sampling_rate numeric(5,4) DEFAULT 1.0 NOT NULL,
    settings_json jsonb DEFAULT '{"theme": {"position": "right", "accent_color": "#00B2FF"}, "launcher": {"call_text": "Позвонить", "primary_text": "Написать"}, "monitoring": {"mask_text": false, "mask_inputs": true, "privacy_mode": true, "mask_selectors": ["[data-rp-mask]", ".rp-mask"], "block_selectors": ["[data-rp-block]", ".rp-block"], "capture_iframes": false}, "active_invite": {"body": "Помочь на этой странице?", "title": "Оператор на связи", "enabled": true, "call_text": "Позвонить", "accept_text": "Ответить", "decline_text": "Не сейчас", "cooldown_minutes": 10}, "enabled_modes": {"call": true, "chat": true, "monitoring": true}, "call_terms_url": "", "call_consent_text": "Мы обработаем аудио только для соединения с оператором. Запись разговора включается только после отдельного согласия.", "recording_enabled": false, "privacy_policy_url": "", "recording_consent_required": true}'::jsonb NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT site_chat_widgets_position_check CHECK (("position" = ANY (ARRAY['left'::text, 'right'::text]))),
    CONSTRAINT site_chat_widgets_public_key_check CHECK ((btrim(public_key) <> ''::text)),
    CONSTRAINT site_chat_widgets_sampling_rate_check CHECK (((sampling_rate >= (0)::numeric) AND (sampling_rate <= (1)::numeric))),
    CONSTRAINT site_chat_widgets_status_check CHECK ((status = ANY (ARRAY['active'::text, 'paused'::text, 'archived'::text])))
);


--
-- Name: site_cobrowse_frames; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.site_cobrowse_frames (
    frame_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    chat_widget_id uuid NOT NULL,
    visitor_id uuid NOT NULL,
    visit_session_id uuid NOT NULL,
    sequence bigint NOT NULL,
    kind text NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL,
    payload_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    payload_compressed bytea,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT site_cobrowse_frames_kind_check CHECK ((kind = ANY (ARRAY['snapshot'::text, 'mutation'::text, 'cursor'::text, 'touch'::text, 'scroll'::text, 'viewport'::text, 'navigation'::text, 'click'::text, 'diagnostic'::text])))
);


--
-- Name: site_visit_connections; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.site_visit_connections (
    connection_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    chat_widget_id uuid,
    visitor_id uuid,
    visit_session_id uuid,
    role text NOT NULL,
    status text DEFAULT 'connected'::text NOT NULL,
    connected_at timestamp with time zone DEFAULT now() NOT NULL,
    last_heartbeat_at timestamp with time zone DEFAULT now() NOT NULL,
    closed_at timestamp with time zone,
    close_reason text DEFAULT ''::text NOT NULL,
    user_agent_hash text DEFAULT ''::text NOT NULL,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT site_visit_connections_role_check CHECK ((role = ANY (ARRAY['visitor'::text, 'operator'::text]))),
    CONSTRAINT site_visit_connections_status_check CHECK ((status = ANY (ARRAY['connected'::text, 'closed'::text, 'timed_out'::text])))
);


--
-- Name: site_visit_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.site_visit_sessions (
    visit_session_id uuid DEFAULT gen_random_uuid() NOT NULL,
    visitor_id uuid NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    chat_widget_id uuid NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    ended_at timestamp with time zone,
    last_activity_at timestamp with time zone DEFAULT now() NOT NULL,
    current_url text DEFAULT ''::text NOT NULL,
    current_title text DEFAULT ''::text NOT NULL,
    viewport_width integer,
    viewport_height integer,
    scroll_x integer DEFAULT 0 NOT NULL,
    scroll_y integer DEFAULT 0 NOT NULL,
    device_type text DEFAULT 'unknown'::text NOT NULL,
    browser text DEFAULT ''::text NOT NULL,
    os text DEFAULT ''::text NOT NULL,
    referrer text DEFAULT ''::text NOT NULL,
    utm jsonb DEFAULT '{}'::jsonb NOT NULL,
    analytics_payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    monitoring_status text DEFAULT 'idle'::text NOT NULL,
    operator_invited boolean DEFAULT false NOT NULL,
    chat_open boolean DEFAULT false NOT NULL,
    event_counts jsonb DEFAULT '{}'::jsonb NOT NULL,
    visitor_state_snapshot jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    duration_seconds integer DEFAULT 0 NOT NULL,
    last_duration_flush_at timestamp with time zone,
    CONSTRAINT site_visit_sessions_monitoring_check CHECK ((monitoring_status = ANY (ARRAY['idle'::text, 'requested'::text, 'active'::text, 'stopped'::text]))),
    CONSTRAINT site_visit_sessions_status_check CHECK ((status = ANY (ARRAY['active'::text, 'idle'::text, 'ended'::text])))
);


--
-- Name: site_visitor_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.site_visitor_events (
    visitor_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    chat_widget_id uuid NOT NULL,
    visitor_id uuid NOT NULL,
    visit_session_id uuid NOT NULL,
    event_type text NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    priority text DEFAULT 'normal'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT site_visitor_events_priority_check CHECK ((priority = ANY (ARRAY['low'::text, 'normal'::text, 'high'::text]))),
    CONSTRAINT site_visitor_events_type_check CHECK ((btrim(event_type) <> ''::text))
);


--
-- Name: site_visitors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.site_visitors (
    visitor_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    chat_widget_id uuid NOT NULL,
    anonymous_id text NOT NULL,
    client_id uuid,
    first_seen_at timestamp with time zone DEFAULT now() NOT NULL,
    last_seen_at timestamp with time zone DEFAULT now() NOT NULL,
    current_url text DEFAULT ''::text NOT NULL,
    current_title text DEFAULT ''::text NOT NULL,
    referrer text DEFAULT ''::text NOT NULL,
    utm jsonb DEFAULT '{}'::jsonb NOT NULL,
    browser text DEFAULT ''::text NOT NULL,
    browser_version text DEFAULT ''::text NOT NULL,
    os text DEFAULT ''::text NOT NULL,
    device_type text DEFAULT 'unknown'::text NOT NULL,
    screen_width integer,
    screen_height integer,
    language text DEFAULT ''::text NOT NULL,
    timezone text DEFAULT ''::text NOT NULL,
    metrika_client_id text DEFAULT ''::text NOT NULL,
    metrika_user_id text DEFAULT ''::text NOT NULL,
    metrika_session_id text DEFAULT ''::text NOT NULL,
    user_agent_hash text DEFAULT ''::text NOT NULL,
    ip_hash text DEFAULT ''::text NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT site_visitors_anonymous_id_check CHECK ((btrim(anonymous_id) <> ''::text))
);


--
-- Name: task_activity_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_activity_events (
    task_activity_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    task_id uuid,
    actor_user_id uuid,
    event_type text NOT NULL,
    payload_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: task_attachments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_attachments (
    task_attachment_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    client_id uuid,
    task_id uuid NOT NULL,
    uploaded_by_user_id uuid,
    file_name text NOT NULL,
    content_type text DEFAULT 'application/octet-stream'::text NOT NULL,
    size_bytes bigint DEFAULT 0 NOT NULL,
    storage_path text NOT NULL,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: task_boards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_boards (
    task_board_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid,
    client_id uuid,
    name text NOT NULL,
    type text DEFAULT 'kanban'::text NOT NULL,
    task_workflow_id uuid NOT NULL,
    group_by text DEFAULT 'status'::text NOT NULL,
    swimlane_by text DEFAULT 'none'::text NOT NULL,
    filters_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    sort_json jsonb DEFAULT '{"field": "rank", "direction": "asc"}'::jsonb NOT NULL,
    visible_fields_json jsonb DEFAULT '["client", "project", "assignee", "estimate", "spent", "due_date", "labels"]'::jsonb NOT NULL,
    card_layout_json jsonb DEFAULT '{"density": "normal"}'::jsonb NOT NULL,
    created_by_user_id uuid,
    is_default boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT task_boards_group_by_check CHECK ((group_by = ANY (ARRAY['status'::text, 'assignee'::text, 'priority'::text, 'client'::text, 'project'::text, 'epic'::text]))),
    CONSTRAINT task_boards_swimlane_by_check CHECK ((swimlane_by = ANY (ARRAY['none'::text, 'assignee'::text, 'priority'::text, 'client'::text, 'project'::text, 'epic'::text, 'deadline'::text]))),
    CONSTRAINT task_boards_type_check CHECK ((type = ANY (ARRAY['kanban'::text, 'scrum'::text, 'support'::text, 'sales'::text, 'custom'::text])))
);


--
-- Name: task_checklist_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_checklist_items (
    task_checklist_item_id uuid DEFAULT gen_random_uuid() NOT NULL,
    task_id uuid NOT NULL,
    title text NOT NULL,
    is_done boolean DEFAULT false NOT NULL,
    "position" integer DEFAULT 0 NOT NULL,
    completed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: task_comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_comments (
    task_comment_id uuid DEFAULT gen_random_uuid() NOT NULL,
    task_id uuid NOT NULL,
    company_id uuid NOT NULL,
    author_user_id uuid,
    body text NOT NULL,
    visibility text DEFAULT 'internal'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT task_comments_visibility_check CHECK ((visibility = ANY (ARRAY['internal'::text, 'client'::text])))
);


--
-- Name: task_custom_field_values; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_custom_field_values (
    task_custom_field_value_id uuid DEFAULT gen_random_uuid() NOT NULL,
    task_custom_field_id uuid NOT NULL,
    task_id uuid,
    project_id uuid,
    client_id uuid,
    employee_profile_id uuid,
    time_entry_id uuid,
    value_json jsonb DEFAULT 'null'::jsonb NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: task_custom_fields; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_custom_fields (
    task_custom_field_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    name text NOT NULL,
    field_key text NOT NULL,
    type text NOT NULL,
    scope text DEFAULT 'task'::text NOT NULL,
    required boolean DEFAULT false NOT NULL,
    options_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    default_value text,
    "position" integer DEFAULT 0 NOT NULL,
    visible_by_default boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    project_id uuid,
    task_workflow_id uuid,
    applies_to_task_types text[],
    is_system boolean DEFAULT false NOT NULL,
    CONSTRAINT task_custom_fields_scope_check CHECK ((scope = ANY (ARRAY['task'::text, 'project'::text, 'client'::text, 'employee'::text, 'time_entry'::text]))),
    CONSTRAINT task_custom_fields_type_check CHECK ((type = ANY (ARRAY['text'::text, 'number'::text, 'money'::text, 'date'::text, 'datetime'::text, 'select'::text, 'multi_select'::text, 'checkbox'::text, 'user'::text, 'users'::text, 'client'::text, 'project'::text, 'url'::text, 'email'::text, 'phone'::text])))
);


--
-- Name: task_process_memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_process_memberships (
    task_process_membership_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    task_workflow_id uuid NOT NULL,
    user_id uuid NOT NULL,
    task_process_role_id uuid NOT NULL,
    client_id uuid,
    invited_by_user_id uuid,
    invited_at timestamp with time zone,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT task_process_memberships_status_check CHECK ((status = ANY (ARRAY['active'::text, 'invited'::text, 'disabled'::text])))
);


--
-- Name: task_process_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_process_roles (
    task_process_role_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid NOT NULL,
    task_workflow_id uuid NOT NULL,
    role_key text NOT NULL,
    name text NOT NULL,
    description text,
    permissions_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    "position" integer DEFAULT 0 NOT NULL,
    is_system boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT task_process_roles_key_check CHECK ((btrim(role_key) <> ''::text)),
    CONSTRAINT task_process_roles_name_check CHECK ((btrim(name) <> ''::text))
);


--
-- Name: task_status_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_status_history (
    task_status_history_id uuid DEFAULT gen_random_uuid() NOT NULL,
    task_id uuid NOT NULL,
    from_status_id uuid,
    to_status_id uuid NOT NULL,
    changed_by_user_id uuid,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    duration_minutes integer
);


--
-- Name: task_views; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_views (
    task_view_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid,
    user_id uuid,
    name text NOT NULL,
    view_type text DEFAULT 'table'::text NOT NULL,
    settings_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT task_views_view_type_check CHECK ((view_type = ANY (ARRAY['table'::text, 'kanban'::text, 'calendar'::text, 'timeline'::text, 'workload'::text])))
);


--
-- Name: task_watchers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_watchers (
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: task_workflow_statuses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_workflow_statuses (
    task_workflow_status_id uuid DEFAULT gen_random_uuid() NOT NULL,
    task_workflow_id uuid NOT NULL,
    name text NOT NULL,
    category text DEFAULT 'todo'::text NOT NULL,
    color text DEFAULT 'slate'::text NOT NULL,
    "position" integer DEFAULT 0 NOT NULL,
    is_start boolean DEFAULT false NOT NULL,
    is_done boolean DEFAULT false NOT NULL,
    is_archived boolean DEFAULT false NOT NULL,
    wip_limit integer,
    auto_start_timer boolean DEFAULT false NOT NULL,
    auto_stop_timer boolean DEFAULT false NOT NULL,
    requires_assignee boolean DEFAULT false NOT NULL,
    requires_estimate boolean DEFAULT false NOT NULL,
    requires_comment_on_enter boolean DEFAULT false NOT NULL,
    requires_comment_on_exit boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT task_workflow_statuses_category_check CHECK ((category = ANY (ARRAY['backlog'::text, 'todo'::text, 'in_progress'::text, 'review'::text, 'done'::text, 'canceled'::text]))),
    CONSTRAINT task_workflow_statuses_wip_check CHECK (((wip_limit IS NULL) OR (wip_limit > 0)))
);


--
-- Name: task_workflow_transitions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_workflow_transitions (
    task_workflow_transition_id uuid DEFAULT gen_random_uuid() NOT NULL,
    task_workflow_id uuid NOT NULL,
    from_status_id uuid,
    to_status_id uuid NOT NULL,
    name text NOT NULL,
    allowed_roles text[] DEFAULT ARRAY[]::text[] NOT NULL,
    required_fields text[] DEFAULT ARRAY[]::text[] NOT NULL,
    automation_rules jsonb DEFAULT '{}'::jsonb NOT NULL,
    "position" integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: task_workflows; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_workflows (
    task_workflow_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid,
    client_id uuid,
    name text NOT NULL,
    description text,
    scope text DEFAULT 'project'::text NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT task_workflows_scope_check CHECK ((scope = ANY (ARRAY['workspace'::text, 'project'::text, 'task_type'::text, 'client'::text])))
);


--
-- Name: tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tasks (
    task_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    project_id uuid,
    client_id uuid,
    parent_task_id uuid,
    task_key text NOT NULL,
    title text NOT NULL,
    description text,
    type text DEFAULT 'task'::text NOT NULL,
    task_workflow_status_id uuid NOT NULL,
    task_workflow_id uuid NOT NULL,
    priority text DEFAULT 'medium'::text NOT NULL,
    severity text DEFAULT 'medium'::text NOT NULL,
    assignee_user_id uuid,
    reporter_user_id uuid,
    reviewer_user_id uuid,
    owner_user_id uuid,
    created_by_user_id uuid,
    start_date date,
    due_date date,
    completed_at timestamp with time zone,
    archived_at timestamp with time zone,
    estimate_minutes integer,
    spent_minutes_cached integer DEFAULT 0 NOT NULL,
    remaining_minutes integer,
    story_points numeric(6,2),
    "position" integer DEFAULT 0 NOT NULL,
    rank numeric(18,6) DEFAULT 0 NOT NULL,
    is_blocked boolean DEFAULT false NOT NULL,
    blocked_reason text,
    sla_due_at timestamp with time zone,
    client_visible boolean DEFAULT false NOT NULL,
    billable boolean DEFAULT true NOT NULL,
    labels text[] DEFAULT ARRAY[]::text[] NOT NULL,
    comments_count integer DEFAULT 0 NOT NULL,
    attachments_count integer DEFAULT 0 NOT NULL,
    checklist_done_count integer DEFAULT 0 NOT NULL,
    checklist_total_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    search_vector tsvector,
    revision integer DEFAULT 0 NOT NULL,
    CONSTRAINT tasks_estimate_check CHECK (((estimate_minutes IS NULL) OR (estimate_minutes >= 0))),
    CONSTRAINT tasks_priority_check CHECK ((priority = ANY (ARRAY['none'::text, 'low'::text, 'medium'::text, 'high'::text, 'urgent'::text]))),
    CONSTRAINT tasks_revision_check CHECK ((revision >= 0)),
    CONSTRAINT tasks_severity_check CHECK ((severity = ANY (ARRAY['low'::text, 'medium'::text, 'high'::text, 'critical'::text]))),
    CONSTRAINT tasks_spent_check CHECK ((spent_minutes_cached >= 0)),
    CONSTRAINT tasks_type_check CHECK ((type = ANY (ARRAY['task'::text, 'bug'::text, 'feature'::text, 'request'::text, 'idea'::text, 'support'::text, 'content'::text, 'design'::text, 'legal'::text, 'finance'::text, 'hr'::text, 'document'::text])))
);


--
-- Name: team_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.team_members (
    team_id uuid NOT NULL,
    employee_profile_id uuid NOT NULL,
    role text DEFAULT 'member'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT team_members_role_check CHECK ((role = ANY (ARRAY['lead'::text, 'member'::text])))
);


--
-- Name: teams; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.teams (
    team_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    name text NOT NULL,
    type text DEFAULT 'team'::text NOT NULL,
    lead_employee_profile_id uuid,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT teams_type_check CHECK ((type = ANY (ARRAY['department'::text, 'team'::text, 'squad'::text])))
);


--
-- Name: telegram_relay_configs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.telegram_relay_configs (
    action_dependency_id uuid NOT NULL,
    relay_key text NOT NULL,
    public_url text NOT NULL,
    webhook_secret_hash text NOT NULL,
    allowed_updates text[] DEFAULT ARRAY['message'::text, 'edited_message'::text, 'channel_post'::text, 'edited_channel_post'::text, 'business_connection'::text, 'business_message'::text, 'edited_business_message'::text, 'deleted_business_messages'::text, 'guest_message'::text, 'message_reaction'::text, 'message_reaction_count'::text, 'inline_query'::text, 'chosen_inline_result'::text, 'callback_query'::text, 'shipping_query'::text, 'pre_checkout_query'::text, 'purchased_paid_media'::text, 'poll'::text, 'poll_answer'::text, 'my_chat_member'::text, 'chat_member'::text, 'chat_join_request'::text, 'chat_boost'::text, 'removed_chat_boost'::text, 'managed_bot'::text] NOT NULL,
    default_chat_id text,
    country_relays_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    polling_enabled boolean DEFAULT false NOT NULL,
    polling_timeout_seconds integer DEFAULT 50 NOT NULL,
    polling_limit integer DEFAULT 100 NOT NULL,
    polling_next_offset bigint,
    polling_last_update_id bigint,
    polling_last_started_at timestamp with time zone,
    polling_last_ok_at timestamp with time zone,
    polling_locked_until timestamp with time zone,
    polling_last_error text,
    polling_delete_webhook boolean DEFAULT true NOT NULL,
    polling_drop_pending_updates boolean DEFAULT false NOT NULL,
    polling_webhook_deleted_at timestamp with time zone,
    CONSTRAINT telegram_relay_configs_polling_limit_check CHECK (((polling_limit >= 1) AND (polling_limit <= 100))),
    CONSTRAINT telegram_relay_configs_polling_timeout_check CHECK (((polling_timeout_seconds >= 1) AND (polling_timeout_seconds <= 60))),
    CONSTRAINT telegram_relay_configs_status_check CHECK ((status = ANY (ARRAY['active'::text, 'disabled'::text, 'deleted'::text])))
);


--
-- Name: telegram_relay_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.telegram_relay_events (
    telegram_relay_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    action_dependency_id uuid NOT NULL,
    project_id uuid NOT NULL,
    relay_key text NOT NULL,
    update_id bigint,
    update_type text DEFAULT ''::text NOT NULL,
    command text DEFAULT ''::text NOT NULL,
    chat_id text DEFAULT ''::text NOT NULL,
    user_id text DEFAULT ''::text NOT NULL,
    message_id bigint,
    text text DEFAULT ''::text NOT NULL,
    payload_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    normalized_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    raw_body bytea DEFAULT '\x'::bytea NOT NULL,
    headers_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    source_ip inet,
    user_agent text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'received'::text NOT NULL,
    workflow_run_ids_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    rejection_reason text,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    queued_at timestamp with time zone,
    processed_at timestamp with time zone,
    error text,
    CONSTRAINT telegram_relay_events_status_check CHECK ((status = ANY (ARRAY['received'::text, 'rejected'::text, 'duplicate'::text, 'queued'::text, 'processed'::text, 'failed'::text, 'ignored'::text])))
);


--
-- Name: telegram_relay_forwards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.telegram_relay_forwards (
    telegram_relay_forward_id uuid DEFAULT gen_random_uuid() NOT NULL,
    telegram_relay_event_id uuid NOT NULL,
    relay_url text NOT NULL,
    relay_region text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    status_code integer,
    request_headers_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    response_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    error text,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    finished_at timestamp with time zone,
    CONSTRAINT telegram_relay_forwards_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'succeeded'::text, 'failed'::text, 'skipped'::text])))
);


--
-- Name: time_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.time_entries (
    time_entry_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    task_id uuid,
    project_id uuid,
    client_id uuid,
    user_id uuid NOT NULL,
    entry_date date DEFAULT CURRENT_DATE NOT NULL,
    started_at timestamp with time zone,
    ended_at timestamp with time zone,
    duration_minutes integer NOT NULL,
    description text,
    billable boolean DEFAULT true NOT NULL,
    source text DEFAULT 'manual'::text NOT NULL,
    task_workflow_status_id uuid,
    approved_by_user_id uuid,
    approved_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT time_entries_duration_check CHECK ((duration_minutes > 0)),
    CONSTRAINT time_entries_source_check CHECK ((source = ANY (ARRAY['manual'::text, 'timer'::text, 'status_auto'::text, 'import'::text])))
);


--
-- Name: timer_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.timer_sessions (
    timer_session_id uuid DEFAULT gen_random_uuid() NOT NULL,
    company_id uuid NOT NULL,
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    paused_at timestamp with time zone,
    stopped_at timestamp with time zone,
    duration_minutes_cached integer DEFAULT 0 NOT NULL,
    state text DEFAULT 'running'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT timer_sessions_duration_check CHECK ((duration_minutes_cached >= 0)),
    CONSTRAINT timer_sessions_state_check CHECK ((state = ANY (ARRAY['running'::text, 'paused'::text, 'stopped'::text])))
);


--
-- Name: user_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_versions (
    user_version_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    version_no integer NOT NULL,
    operation_type text NOT NULL,
    snapshot jsonb NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by_user_id uuid,
    request_id text,
    CONSTRAINT user_versions_operation_type_check CHECK ((operation_type = ANY (ARRAY['insert'::text, 'update'::text, 'delete'::text])))
);


--
-- Name: user_workspace_contexts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_workspace_contexts (
    user_id uuid NOT NULL,
    active_company_id uuid,
    active_project_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    user_id uuid DEFAULT gen_random_uuid() NOT NULL,
    email public.citext,
    password_hash text,
    first_name text,
    is_activated boolean DEFAULT false NOT NULL,
    activation_token_hash text,
    activation_token_expires_at timestamp with time zone,
    password_reset_token_hash text,
    password_reset_token_expires_at timestamp with time zone,
    email_change_token_hash text,
    email_change_token_expires_at timestamp with time zone,
    pending_email public.citext,
    current_company_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    display_name text,
    CONSTRAINT users_email_or_password_material_check CHECK (((email IS NULL) OR (password_hash IS NOT NULL) OR (activation_token_hash IS NOT NULL)))
);


--
-- Name: webhook_endpoints; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.webhook_endpoints (
    webhook_endpoint_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    workflow_id uuid NOT NULL,
    workflow_version_id uuid NOT NULL,
    name text NOT NULL,
    endpoint_key text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    auth_type text DEFAULT 'project_secret'::text NOT NULL,
    secret_hash text,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    secret_ciphertext bytea,
    previous_secret_hash text,
    previous_secret_ciphertext bytea,
    previous_secret_expires_at timestamp with time zone,
    replay_window_seconds integer DEFAULT 300 NOT NULL,
    CONSTRAINT webhook_endpoints_auth_type_check CHECK ((auth_type = ANY (ARRAY['none'::text, 'project_secret'::text, 'hmac'::text, 'bearer'::text]))),
    CONSTRAINT webhook_endpoints_replay_window_seconds_check CHECK (((replay_window_seconds >= 30) AND (replay_window_seconds <= 86400))),
    CONSTRAINT webhook_endpoints_status_check CHECK ((status = ANY (ARRAY['active'::text, 'paused'::text, 'disabled'::text, 'deleted'::text])))
);


--
-- Name: webhook_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.webhook_events (
    webhook_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    webhook_endpoint_id uuid NOT NULL,
    workflow_id uuid NOT NULL,
    workflow_version_id uuid NOT NULL,
    workflow_run_id uuid,
    method text NOT NULL,
    path text NOT NULL,
    headers_json jsonb,
    query_json jsonb,
    payload_json jsonb,
    raw_body bytea,
    content_type text,
    source_ip inet,
    user_agent text,
    status text DEFAULT 'received'::text NOT NULL,
    rejection_reason text,
    idempotency_key text,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    queued_at timestamp with time zone,
    processed_at timestamp with time zone,
    error text,
    event_id text,
    received_timestamp timestamp with time zone,
    body_sha256 text,
    auth_scheme text,
    signature_verified boolean DEFAULT false NOT NULL,
    CONSTRAINT webhook_events_status_check CHECK ((status = ANY (ARRAY['received'::text, 'rejected'::text, 'duplicate'::text, 'queued'::text, 'processing'::text, 'processed'::text, 'failed'::text, 'ignored'::text])))
);


--
-- Name: workflow_action_dead_letters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.workflow_action_dead_letters (
    dead_letter_id uuid DEFAULT gen_random_uuid() NOT NULL,
    workflow_run_id uuid NOT NULL,
    action_run_id uuid NOT NULL,
    action_key text NOT NULL,
    action_type_key text NOT NULL,
    retry_class text NOT NULL,
    redacted_reason text NOT NULL,
    owner text NOT NULL,
    replayable boolean DEFAULT false NOT NULL,
    retry_policy_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    replayed_at timestamp with time zone,
    replay_workflow_run_id uuid,
    CONSTRAINT workflow_action_dead_letters_retry_class_check CHECK ((retry_class = ANY (ARRAY['validation'::text, 'auth'::text, 'permanent'::text, 'transient'::text, 'rate_limited'::text, 'timeout_unknown'::text, 'canceled'::text, 'unknown_outcome'::text])))
);


--
-- Name: workflow_client_payment_due_state; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.workflow_client_payment_due_state (
    workflow_id uuid NOT NULL,
    action_key text NOT NULL,
    workflow_version_id uuid NOT NULL,
    project_id uuid NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    remind_every_days integer DEFAULT 7 NOT NULL,
    days_before_due integer DEFAULT 7 NOT NULL,
    timezone text DEFAULT 'Europe/Moscow'::text NOT NULL,
    last_scan_at timestamp with time zone,
    last_workflow_run_id uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: workflow_runs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.workflow_runs (
    workflow_run_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    workflow_id uuid NOT NULL,
    workflow_version_id uuid NOT NULL,
    source_type text NOT NULL,
    source_webhook_event_id uuid,
    replayed_from_workflow_run_id uuid,
    status text DEFAULT 'pending'::text NOT NULL,
    input_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    output_json jsonb,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    error text,
    run_request_version text DEFAULT 'run-request.v1'::text NOT NULL,
    run_request_idempotency_key text DEFAULT ''::text NOT NULL,
    run_request_source_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    run_claim_owner text DEFAULT ''::text NOT NULL,
    run_lease_expires_at timestamp with time zone,
    run_heartbeat_at timestamp with time zone,
    run_fencing_token bigint DEFAULT 0 NOT NULL,
    run_claim_count integer DEFAULT 0 NOT NULL,
    state_machine_version text DEFAULT 'workflow-run-state.v1'::text NOT NULL,
    terminal_reason text DEFAULT ''::text NOT NULL,
    cancel_reason text DEFAULT ''::text NOT NULL,
    cancel_requested_by_user_id uuid,
    cancelled_at timestamp with time zone,
    waiting_deadline_at timestamp with time zone,
    waiting_reason text DEFAULT ''::text NOT NULL,
    wake_reason text DEFAULT ''::text NOT NULL,
    last_state_transition_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT workflow_runs_run_claim_count_check CHECK ((run_claim_count >= 0)),
    CONSTRAINT workflow_runs_run_fencing_token_check CHECK ((run_fencing_token >= 0)),
    CONSTRAINT workflow_runs_source_type_check CHECK ((source_type = ANY (ARRAY['webhook'::text, 'manual'::text, 'schedule'::text, 'desktop_message'::text, 'replay'::text, 'test'::text, 'telegram_relay'::text, 'client_payment_due'::text, 'module_trigger'::text, 'task_event'::text]))),
    CONSTRAINT workflow_runs_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'running'::text, 'succeeded'::text, 'failed'::text, 'cancelled'::text, 'waiting'::text])))
);


--
-- Name: workflow_schedule_state; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.workflow_schedule_state (
    workflow_id uuid NOT NULL,
    action_key text NOT NULL,
    workflow_version_id uuid NOT NULL,
    schedule_key text NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    interval_seconds integer DEFAULT 3600 NOT NULL,
    next_run_at timestamp with time zone DEFAULT now() NOT NULL,
    last_run_at timestamp with time zone,
    last_workflow_run_id uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: workflow_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.workflow_versions (
    workflow_version_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    workflow_id uuid NOT NULL,
    version_no integer NOT NULL,
    status text DEFAULT 'published'::text NOT NULL,
    definition_json jsonb NOT NULL,
    dsl_yaml text,
    graph_json jsonb,
    execution_plan_json jsonb,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    published_at timestamp with time zone DEFAULT now() NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    restored_from_workflow_version_id uuid,
    CONSTRAINT workflow_versions_status_check CHECK ((status = ANY (ARRAY['published'::text, 'deprecated'::text, 'test'::text])))
);


--
-- Name: workflows; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.workflows (
    workflow_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL,
    draft_definition_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    draft_dsl_yaml text,
    draft_graph_json jsonb,
    current_workflow_version_id uuid,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    test_input_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT workflows_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'active'::text, 'paused'::text, 'archived'::text])))
);


--
-- Name: ai_thread_scopes ai_thread_scopes_pkey; Type: CONSTRAINT; Schema: ai_agent_checkpoint; Owner: -
--

ALTER TABLE ONLY ai_agent_checkpoint.ai_thread_scopes
    ADD CONSTRAINT ai_thread_scopes_pkey PRIMARY KEY (checkpoint_namespace, checkpoint_thread_id);


--
-- Name: action_attempts action_attempts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_attempts
    ADD CONSTRAINT action_attempts_pkey PRIMARY KEY (action_attempt_id);


--
-- Name: action_dependencies action_dependencies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_dependencies
    ADD CONSTRAINT action_dependencies_pkey PRIMARY KEY (action_dependency_id);


--
-- Name: action_dependency_types action_dependency_types_dependency_type_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_dependency_types
    ADD CONSTRAINT action_dependency_types_dependency_type_key_key UNIQUE (dependency_type_key);


--
-- Name: action_dependency_types action_dependency_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_dependency_types
    ADD CONSTRAINT action_dependency_types_pkey PRIMARY KEY (action_dependency_type_id);


--
-- Name: action_runs action_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_runs
    ADD CONSTRAINT action_runs_pkey PRIMARY KEY (action_run_id);


--
-- Name: action_types action_types_action_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_types
    ADD CONSTRAINT action_types_action_key_key UNIQUE (action_key);


--
-- Name: action_types action_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_types
    ADD CONSTRAINT action_types_pkey PRIMARY KEY (action_type_id);


--
-- Name: ai_agent_audit_events ai_agent_audit_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_audit_events
    ADD CONSTRAINT ai_agent_audit_events_pkey PRIMARY KEY (ai_agent_audit_event_id);


--
-- Name: ai_agent_events ai_agent_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_events
    ADD CONSTRAINT ai_agent_events_pkey PRIMARY KEY (ai_agent_event_id);


--
-- Name: ai_agent_runs ai_agent_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_runs
    ADD CONSTRAINT ai_agent_runs_pkey PRIMARY KEY (ai_agent_run_id);


--
-- Name: ai_channel_bindings ai_channel_bindings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_channel_bindings
    ADD CONSTRAINT ai_channel_bindings_pkey PRIMARY KEY (ai_channel_binding_id);


--
-- Name: ai_chat_cards ai_chat_cards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_cards
    ADD CONSTRAINT ai_chat_cards_pkey PRIMARY KEY (ai_chat_card_id);


--
-- Name: ai_chat_messages ai_chat_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_messages
    ADD CONSTRAINT ai_chat_messages_pkey PRIMARY KEY (ai_chat_message_id);


--
-- Name: ai_chat_sessions ai_chat_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_sessions
    ADD CONSTRAINT ai_chat_sessions_pkey PRIMARY KEY (ai_chat_session_id);


--
-- Name: ai_model_catalog_cache ai_model_catalog_cache_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_model_catalog_cache
    ADD CONSTRAINT ai_model_catalog_cache_pkey PRIMARY KEY (ai_model_catalog_cache_id);


--
-- Name: ai_model_evaluations ai_model_evaluations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_model_evaluations
    ADD CONSTRAINT ai_model_evaluations_pkey PRIMARY KEY (ai_model_evaluation_id);


--
-- Name: ai_patch_proposals ai_patch_proposals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_patch_proposals
    ADD CONSTRAINT ai_patch_proposals_pkey PRIMARY KEY (ai_patch_proposal_id);


--
-- Name: ai_prompt_snapshots ai_prompt_snapshots_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_prompt_snapshots
    ADD CONSTRAINT ai_prompt_snapshots_pkey PRIMARY KEY (ai_prompt_snapshot_id);


--
-- Name: ai_session_context_frames ai_session_context_frames_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_session_context_frames
    ADD CONSTRAINT ai_session_context_frames_pkey PRIMARY KEY (ai_session_context_frame_id);


--
-- Name: ai_tool_call_edits ai_tool_call_edits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_call_edits
    ADD CONSTRAINT ai_tool_call_edits_pkey PRIMARY KEY (ai_tool_call_edit_id);


--
-- Name: ai_tool_calls ai_tool_calls_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_calls
    ADD CONSTRAINT ai_tool_calls_pkey PRIMARY KEY (ai_tool_call_id);


--
-- Name: ai_tool_idempotency ai_tool_idempotency_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_idempotency
    ADD CONSTRAINT ai_tool_idempotency_pkey PRIMARY KEY (ai_tool_idempotency_id);


--
-- Name: ai_tool_plan_steps ai_tool_plan_steps_ai_tool_plan_id_step_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_plan_steps
    ADD CONSTRAINT ai_tool_plan_steps_ai_tool_plan_id_step_id_key UNIQUE (ai_tool_plan_id, step_id);


--
-- Name: ai_tool_plan_steps ai_tool_plan_steps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_plan_steps
    ADD CONSTRAINT ai_tool_plan_steps_pkey PRIMARY KEY (ai_tool_plan_step_id);


--
-- Name: ai_tool_plans ai_tool_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_plans
    ADD CONSTRAINT ai_tool_plans_pkey PRIMARY KEY (ai_tool_plan_id);


--
-- Name: ai_usage_events ai_usage_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_usage_events
    ADD CONSTRAINT ai_usage_events_pkey PRIMARY KEY (ai_usage_event_id);


--
-- Name: api_key_grants api_key_grants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_key_grants
    ADD CONSTRAINT api_key_grants_pkey PRIMARY KEY (api_key_grant_id);


--
-- Name: api_keys api_keys_key_prefix_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_key_prefix_key UNIQUE (key_prefix);


--
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (api_key_id);


--
-- Name: audit_events audit_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_events
    ADD CONSTRAINT audit_events_pkey PRIMARY KEY (audit_event_id);


--
-- Name: auth_events auth_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_events
    ADD CONSTRAINT auth_events_pkey PRIMARY KEY (auth_event_id);


--
-- Name: auth_identities auth_identities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_identities
    ADD CONSTRAINT auth_identities_pkey PRIMARY KEY (auth_identity_id);


--
-- Name: auth_identity_versions auth_identity_versions_entity_version_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_identity_versions
    ADD CONSTRAINT auth_identity_versions_entity_version_unique UNIQUE (auth_identity_id, version_no);


--
-- Name: auth_identity_versions auth_identity_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_identity_versions
    ADD CONSTRAINT auth_identity_versions_pkey PRIMARY KEY (auth_identity_version_id);


--
-- Name: automation_rules automation_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.automation_rules
    ADD CONSTRAINT automation_rules_pkey PRIMARY KEY (automation_rule_id);


--
-- Name: bot_sessions bot_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_sessions
    ADD CONSTRAINT bot_sessions_pkey PRIMARY KEY (bot_session_id);


--
-- Name: client_contacts client_contacts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_contacts
    ADD CONSTRAINT client_contacts_pkey PRIMARY KEY (client_contact_id);


--
-- Name: client_projects client_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_projects
    ADD CONSTRAINT client_projects_pkey PRIMARY KEY (client_id, project_id);


--
-- Name: client_responsibles client_responsibles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_responsibles
    ADD CONSTRAINT client_responsibles_pkey PRIMARY KEY (client_responsible_id);


--
-- Name: client_service_types client_service_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_service_types
    ADD CONSTRAINT client_service_types_pkey PRIMARY KEY (client_service_type_id);


--
-- Name: client_services client_services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_services
    ADD CONSTRAINT client_services_pkey PRIMARY KEY (client_service_id);


--
-- Name: clients clients_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clients
    ADD CONSTRAINT clients_pkey PRIMARY KEY (client_id);


--
-- Name: companies companies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.companies
    ADD CONSTRAINT companies_pkey PRIMARY KEY (company_id);


--
-- Name: company_legal_entities company_legal_entities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_legal_entities
    ADD CONSTRAINT company_legal_entities_pkey PRIMARY KEY (company_legal_entity_id);


--
-- Name: company_membership_versions company_membership_versions_entity_version_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_membership_versions
    ADD CONSTRAINT company_membership_versions_entity_version_unique UNIQUE (company_id, user_id, version_no);


--
-- Name: company_membership_versions company_membership_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_membership_versions
    ADD CONSTRAINT company_membership_versions_pkey PRIMARY KEY (company_membership_version_id);


--
-- Name: company_memberships company_memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_memberships
    ADD CONSTRAINT company_memberships_pkey PRIMARY KEY (company_id, user_id);


--
-- Name: company_versions company_versions_company_version_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_versions
    ADD CONSTRAINT company_versions_company_version_unique UNIQUE (company_id, version_no);


--
-- Name: company_versions company_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_versions
    ADD CONSTRAINT company_versions_pkey PRIMARY KEY (company_version_id);


--
-- Name: desktop_auth_requests desktop_auth_requests_device_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_auth_requests
    ADD CONSTRAINT desktop_auth_requests_device_code_key UNIQUE (device_code);


--
-- Name: desktop_auth_requests desktop_auth_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_auth_requests
    ADD CONSTRAINT desktop_auth_requests_pkey PRIMARY KEY (desktop_auth_request_id);


--
-- Name: desktop_auth_requests desktop_auth_requests_user_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_auth_requests
    ADD CONSTRAINT desktop_auth_requests_user_code_key UNIQUE (user_code);


--
-- Name: desktop_devices desktop_devices_device_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_devices
    ADD CONSTRAINT desktop_devices_device_key_key UNIQUE (device_key);


--
-- Name: desktop_devices desktop_devices_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_devices
    ADD CONSTRAINT desktop_devices_pkey PRIMARY KEY (desktop_device_id);


--
-- Name: desktop_messages desktop_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_messages
    ADD CONSTRAINT desktop_messages_pkey PRIMARY KEY (desktop_message_id);


--
-- Name: desktop_tokens desktop_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_tokens
    ADD CONSTRAINT desktop_tokens_pkey PRIMARY KEY (desktop_token_id);


--
-- Name: desktop_tokens desktop_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_tokens
    ADD CONSTRAINT desktop_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: device_credential_versions device_credential_versions_entity_version_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_credential_versions
    ADD CONSTRAINT device_credential_versions_entity_version_unique UNIQUE (device_credential_id, version_no);


--
-- Name: device_credential_versions device_credential_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_credential_versions
    ADD CONSTRAINT device_credential_versions_pkey PRIMARY KEY (device_credential_version_id);


--
-- Name: device_credentials device_credentials_credential_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_credentials
    ADD CONSTRAINT device_credentials_credential_hash_key UNIQUE (credential_hash);


--
-- Name: device_credentials device_credentials_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_credentials
    ADD CONSTRAINT device_credentials_pkey PRIMARY KEY (device_credential_id);


--
-- Name: effect_ledger effect_ledger_idempotency_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.effect_ledger
    ADD CONSTRAINT effect_ledger_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: effect_ledger effect_ledger_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.effect_ledger
    ADD CONSTRAINT effect_ledger_pkey PRIMARY KEY (effect_id);


--
-- Name: email_templates email_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_templates
    ADD CONSTRAINT email_templates_pkey PRIMARY KEY (email_template_id);


--
-- Name: employee_profiles employee_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_profiles
    ADD CONSTRAINT employee_profiles_pkey PRIMARY KEY (employee_profile_id);


--
-- Name: integration_doc_chunks integration_doc_chunks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.integration_doc_chunks
    ADD CONSTRAINT integration_doc_chunks_pkey PRIMARY KEY (integration_doc_chunk_id);


--
-- Name: integration_docs integration_docs_doc_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.integration_docs
    ADD CONSTRAINT integration_docs_doc_key_key UNIQUE (doc_key);


--
-- Name: integration_docs integration_docs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.integration_docs
    ADD CONSTRAINT integration_docs_pkey PRIMARY KEY (integration_doc_id);


--
-- Name: integration_docs integration_docs_public_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.integration_docs
    ADD CONSTRAINT integration_docs_public_slug_key UNIQUE (public_slug);


--
-- Name: message_templates message_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message_templates
    ADD CONSTRAINT message_templates_pkey PRIMARY KEY (message_template_id);


--
-- Name: module_channel_bindings module_channel_bindings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_channel_bindings
    ADD CONSTRAINT module_channel_bindings_pkey PRIMARY KEY (module_channel_binding_id);


--
-- Name: module_workflow_triggers module_workflow_triggers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_workflow_triggers
    ADD CONSTRAINT module_workflow_triggers_pkey PRIMARY KEY (module_workflow_trigger_id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (notification_id);


--
-- Name: outbox_events outbox_events_idempotency_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.outbox_events
    ADD CONSTRAINT outbox_events_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: outbox_events outbox_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.outbox_events
    ADD CONSTRAINT outbox_events_pkey PRIMARY KEY (outbox_event_id);


--
-- Name: payment_action_documents payment_action_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_action_documents
    ADD CONSTRAINT payment_action_documents_pkey PRIMARY KEY (payment_action_document_id);


--
-- Name: payment_action_items payment_action_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_action_items
    ADD CONSTRAINT payment_action_items_pkey PRIMARY KEY (payment_action_item_id);


--
-- Name: payment_actions payment_actions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_actions
    ADD CONSTRAINT payment_actions_pkey PRIMARY KEY (payment_action_id);


--
-- Name: payment_events payment_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_events
    ADD CONSTRAINT payment_events_pkey PRIMARY KEY (payment_event_id);


--
-- Name: payment_facts payment_facts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_facts
    ADD CONSTRAINT payment_facts_pkey PRIMARY KEY (payment_fact_id);


--
-- Name: project_ai_employees project_ai_employees_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_ai_employees
    ADD CONSTRAINT project_ai_employees_pkey PRIMARY KEY (project_ai_employee_id);


--
-- Name: project_member_versions project_member_versions_entity_version_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_member_versions
    ADD CONSTRAINT project_member_versions_entity_version_unique UNIQUE (project_id, user_id, version_no);


--
-- Name: project_member_versions project_member_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_member_versions
    ADD CONSTRAINT project_member_versions_pkey PRIMARY KEY (project_member_version_id);


--
-- Name: project_members project_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_pkey PRIMARY KEY (project_id, user_id);


--
-- Name: project_versions project_versions_entity_version_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_versions
    ADD CONSTRAINT project_versions_entity_version_unique UNIQUE (project_id, version_no);


--
-- Name: project_versions project_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_versions
    ADD CONSTRAINT project_versions_pkey PRIMARY KEY (project_version_id);


--
-- Name: projects projects_hook_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_hook_key_key UNIQUE (hook_key);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (project_id);


--
-- Name: provider_operation_effects provider_operation_effects_identity_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.provider_operation_effects
    ADD CONSTRAINT provider_operation_effects_identity_unique UNIQUE (action_dependency_id, provider, operation_key);


--
-- Name: provider_operation_effects provider_operation_effects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.provider_operation_effects
    ADD CONSTRAINT provider_operation_effects_pkey PRIMARY KEY (provider_operation_effect_id);


--
-- Name: run_events run_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.run_events
    ADD CONSTRAINT run_events_pkey PRIMARY KEY (run_event_id);


--
-- Name: secrets secrets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.secrets
    ADD CONSTRAINT secrets_pkey PRIMARY KEY (secret_id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (session_id);


--
-- Name: sessions sessions_refresh_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_refresh_token_hash_key UNIQUE (refresh_token_hash);


--
-- Name: sessions sessions_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_token_hash_key UNIQUE (session_token_hash);


--
-- Name: site_call_sessions site_call_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_call_sessions
    ADD CONSTRAINT site_call_sessions_pkey PRIMARY KEY (call_session_id);


--
-- Name: site_chat_attachments site_chat_attachments_download_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_attachments
    ADD CONSTRAINT site_chat_attachments_download_token_key UNIQUE (download_token);


--
-- Name: site_chat_attachments site_chat_attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_attachments
    ADD CONSTRAINT site_chat_attachments_pkey PRIMARY KEY (attachment_id);


--
-- Name: site_chat_conversations site_chat_conversations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_conversations
    ADD CONSTRAINT site_chat_conversations_pkey PRIMARY KEY (conversation_id);


--
-- Name: site_chat_messages site_chat_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_messages
    ADD CONSTRAINT site_chat_messages_pkey PRIMARY KEY (message_id);


--
-- Name: site_chat_widgets site_chat_widgets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_widgets
    ADD CONSTRAINT site_chat_widgets_pkey PRIMARY KEY (chat_widget_id);


--
-- Name: site_chat_widgets site_chat_widgets_public_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_widgets
    ADD CONSTRAINT site_chat_widgets_public_key_key UNIQUE (public_key);


--
-- Name: site_cobrowse_frames site_cobrowse_frames_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_cobrowse_frames
    ADD CONSTRAINT site_cobrowse_frames_pkey PRIMARY KEY (frame_id);


--
-- Name: site_visit_connections site_visit_connections_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visit_connections
    ADD CONSTRAINT site_visit_connections_pkey PRIMARY KEY (connection_id);


--
-- Name: site_visit_sessions site_visit_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visit_sessions
    ADD CONSTRAINT site_visit_sessions_pkey PRIMARY KEY (visit_session_id);


--
-- Name: site_visitor_events site_visitor_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visitor_events
    ADD CONSTRAINT site_visitor_events_pkey PRIMARY KEY (visitor_event_id);


--
-- Name: site_visitors site_visitors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visitors
    ADD CONSTRAINT site_visitors_pkey PRIMARY KEY (visitor_id);


--
-- Name: task_activity_events task_activity_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_activity_events
    ADD CONSTRAINT task_activity_events_pkey PRIMARY KEY (task_activity_event_id);


--
-- Name: task_attachments task_attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_pkey PRIMARY KEY (task_attachment_id);


--
-- Name: task_boards task_boards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_boards
    ADD CONSTRAINT task_boards_pkey PRIMARY KEY (task_board_id);


--
-- Name: task_checklist_items task_checklist_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_checklist_items
    ADD CONSTRAINT task_checklist_items_pkey PRIMARY KEY (task_checklist_item_id);


--
-- Name: task_comments task_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_pkey PRIMARY KEY (task_comment_id);


--
-- Name: task_custom_field_values task_custom_field_values_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_custom_field_values
    ADD CONSTRAINT task_custom_field_values_pkey PRIMARY KEY (task_custom_field_value_id);


--
-- Name: task_custom_fields task_custom_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_custom_fields
    ADD CONSTRAINT task_custom_fields_pkey PRIMARY KEY (task_custom_field_id);


--
-- Name: task_process_memberships task_process_memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_process_memberships
    ADD CONSTRAINT task_process_memberships_pkey PRIMARY KEY (task_process_membership_id);


--
-- Name: task_process_roles task_process_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_process_roles
    ADD CONSTRAINT task_process_roles_pkey PRIMARY KEY (task_process_role_id);


--
-- Name: task_status_history task_status_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_status_history
    ADD CONSTRAINT task_status_history_pkey PRIMARY KEY (task_status_history_id);


--
-- Name: task_views task_views_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_views
    ADD CONSTRAINT task_views_pkey PRIMARY KEY (task_view_id);


--
-- Name: task_watchers task_watchers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_watchers
    ADD CONSTRAINT task_watchers_pkey PRIMARY KEY (task_id, user_id);


--
-- Name: task_workflow_statuses task_workflow_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_workflow_statuses
    ADD CONSTRAINT task_workflow_statuses_pkey PRIMARY KEY (task_workflow_status_id);


--
-- Name: task_workflow_transitions task_workflow_transitions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_workflow_transitions
    ADD CONSTRAINT task_workflow_transitions_pkey PRIMARY KEY (task_workflow_transition_id);


--
-- Name: task_workflows task_workflows_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_workflows
    ADD CONSTRAINT task_workflows_pkey PRIMARY KEY (task_workflow_id);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (task_id);


--
-- Name: team_members team_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_pkey PRIMARY KEY (team_id, employee_profile_id);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (team_id);


--
-- Name: telegram_relay_configs telegram_relay_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.telegram_relay_configs
    ADD CONSTRAINT telegram_relay_configs_pkey PRIMARY KEY (action_dependency_id);


--
-- Name: telegram_relay_configs telegram_relay_configs_relay_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.telegram_relay_configs
    ADD CONSTRAINT telegram_relay_configs_relay_key_key UNIQUE (relay_key);


--
-- Name: telegram_relay_events telegram_relay_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.telegram_relay_events
    ADD CONSTRAINT telegram_relay_events_pkey PRIMARY KEY (telegram_relay_event_id);


--
-- Name: telegram_relay_forwards telegram_relay_forwards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.telegram_relay_forwards
    ADD CONSTRAINT telegram_relay_forwards_pkey PRIMARY KEY (telegram_relay_forward_id);


--
-- Name: time_entries time_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_pkey PRIMARY KEY (time_entry_id);


--
-- Name: timer_sessions timer_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timer_sessions
    ADD CONSTRAINT timer_sessions_pkey PRIMARY KEY (timer_session_id);


--
-- Name: action_attempts uidx_action_attempts_action_run_attempt; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_attempts
    ADD CONSTRAINT uidx_action_attempts_action_run_attempt UNIQUE (action_run_id, attempt_no);


--
-- Name: action_runs uidx_action_runs_workflow_run_action_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_runs
    ADD CONSTRAINT uidx_action_runs_workflow_run_action_key UNIQUE (workflow_run_id, action_key);


--
-- Name: ai_channel_bindings uidx_ai_channel_bindings_employee_channel; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_channel_bindings
    ADD CONSTRAINT uidx_ai_channel_bindings_employee_channel UNIQUE (project_ai_employee_id, channel_key);


--
-- Name: ai_model_catalog_cache uidx_ai_model_catalog_cache_dependency; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_model_catalog_cache
    ADD CONSTRAINT uidx_ai_model_catalog_cache_dependency UNIQUE (action_dependency_id);


--
-- Name: bot_sessions uidx_bot_sessions_identity; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_sessions
    ADD CONSTRAINT uidx_bot_sessions_identity UNIQUE (project_id, action_dependency_id, relay_key, chat_id, user_id);


--
-- Name: email_templates uidx_email_templates_project_id_slug; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_templates
    ADD CONSTRAINT uidx_email_templates_project_id_slug UNIQUE (project_id, slug);


--
-- Name: integration_doc_chunks uidx_integration_doc_chunks_doc_index; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.integration_doc_chunks
    ADD CONSTRAINT uidx_integration_doc_chunks_doc_index UNIQUE (integration_doc_id, chunk_index);


--
-- Name: message_templates uidx_message_templates_project_provider_slug; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message_templates
    ADD CONSTRAINT uidx_message_templates_project_provider_slug UNIQUE (project_id, provider_key, slug);


--
-- Name: project_ai_employees uidx_project_ai_employees_project_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_ai_employees
    ADD CONSTRAINT uidx_project_ai_employees_project_key UNIQUE (project_id, employee_key);


--
-- Name: projects uidx_projects_company_id_slug; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT uidx_projects_company_id_slug UNIQUE (company_id, slug);


--
-- Name: task_custom_fields uidx_task_custom_fields_company_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_custom_fields
    ADD CONSTRAINT uidx_task_custom_fields_company_key UNIQUE (company_id, field_key);


--
-- Name: teams uidx_teams_company_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT uidx_teams_company_name UNIQUE (company_id, name);


--
-- Name: workflow_versions uidx_workflow_versions_workflow_id_version_no; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_versions
    ADD CONSTRAINT uidx_workflow_versions_workflow_id_version_no UNIQUE (workflow_id, version_no);


--
-- Name: workflows uidx_workflows_project_id_slug; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT uidx_workflows_project_id_slug UNIQUE (project_id, slug);


--
-- Name: user_versions user_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_versions
    ADD CONSTRAINT user_versions_pkey PRIMARY KEY (user_version_id);


--
-- Name: user_versions user_versions_user_version_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_versions
    ADD CONSTRAINT user_versions_user_version_unique UNIQUE (user_id, version_no);


--
-- Name: user_workspace_contexts user_workspace_contexts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_workspace_contexts
    ADD CONSTRAINT user_workspace_contexts_pkey PRIMARY KEY (user_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- Name: webhook_endpoints webhook_endpoints_endpoint_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_endpoints
    ADD CONSTRAINT webhook_endpoints_endpoint_key_key UNIQUE (endpoint_key);


--
-- Name: webhook_endpoints webhook_endpoints_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_endpoints
    ADD CONSTRAINT webhook_endpoints_pkey PRIMARY KEY (webhook_endpoint_id);


--
-- Name: webhook_events webhook_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_pkey PRIMARY KEY (webhook_event_id);


--
-- Name: workflow_action_dead_letters workflow_action_dead_letters_action_run_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_action_dead_letters
    ADD CONSTRAINT workflow_action_dead_letters_action_run_unique UNIQUE (action_run_id);


--
-- Name: workflow_action_dead_letters workflow_action_dead_letters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_action_dead_letters
    ADD CONSTRAINT workflow_action_dead_letters_pkey PRIMARY KEY (dead_letter_id);


--
-- Name: workflow_client_payment_due_state workflow_client_payment_due_state_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_client_payment_due_state
    ADD CONSTRAINT workflow_client_payment_due_state_pkey PRIMARY KEY (workflow_id, action_key);


--
-- Name: workflow_runs workflow_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_runs
    ADD CONSTRAINT workflow_runs_pkey PRIMARY KEY (workflow_run_id);


--
-- Name: workflow_schedule_state workflow_schedule_state_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_schedule_state
    ADD CONSTRAINT workflow_schedule_state_pkey PRIMARY KEY (workflow_id, action_key);


--
-- Name: workflow_versions workflow_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_versions
    ADD CONSTRAINT workflow_versions_pkey PRIMARY KEY (workflow_version_id);


--
-- Name: workflows workflows_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT workflows_pkey PRIMARY KEY (workflow_id);


--
-- Name: idx_ai_thread_scopes_cleanup; Type: INDEX; Schema: ai_agent_checkpoint; Owner: -
--

CREATE INDEX idx_ai_thread_scopes_cleanup ON ai_agent_checkpoint.ai_thread_scopes USING btree (expires_at, status);


--
-- Name: idx_action_attempts_action_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_action_attempts_action_run_id ON public.action_attempts USING btree (action_run_id);


--
-- Name: idx_action_dependencies_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_action_dependencies_company_id ON public.action_dependencies USING btree (company_id);


--
-- Name: idx_action_dependencies_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_action_dependencies_project_id ON public.action_dependencies USING btree (project_id);


--
-- Name: idx_action_dependencies_project_id_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_action_dependencies_project_id_type ON public.action_dependencies USING btree (project_id, action_dependency_type_id);


--
-- Name: idx_action_dependencies_scope; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_action_dependencies_scope ON public.action_dependencies USING btree (scope_kind, company_id, project_id);


--
-- Name: idx_action_runs_dead_lettered; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_action_runs_dead_lettered ON public.action_runs USING btree (dead_lettered_at DESC, retry_class, action_type_key) WHERE (dead_lettered_at IS NOT NULL);


--
-- Name: idx_action_runs_retrying_next_retry; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_action_runs_retrying_next_retry ON public.action_runs USING btree (next_retry_at, action_type_key) WHERE (status = 'retrying'::text);


--
-- Name: idx_action_runs_waiting_deadline; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_action_runs_waiting_deadline ON public.action_runs USING btree (waiting_deadline_at, action_run_id) WHERE ((status = 'waiting'::text) AND (waiting_deadline_at IS NOT NULL));


--
-- Name: idx_action_runs_workflow_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_action_runs_workflow_run_id ON public.action_runs USING btree (workflow_run_id);


--
-- Name: idx_ai_channel_bindings_project; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_channel_bindings_project ON public.ai_channel_bindings USING btree (project_id, channel_key, status);


--
-- Name: idx_ai_chat_messages_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_chat_messages_session_id ON public.ai_chat_messages USING btree (ai_chat_session_id, created_at);


--
-- Name: idx_ai_chat_sessions_active_card; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_chat_sessions_active_card ON public.ai_chat_sessions USING btree (active_card_id) WHERE (active_card_id <> ''::text);


--
-- Name: idx_ai_chat_sessions_active_tool_call; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_chat_sessions_active_tool_call ON public.ai_chat_sessions USING btree (active_tool_call_id) WHERE (active_tool_call_id IS NOT NULL);


--
-- Name: idx_ai_chat_sessions_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_chat_sessions_created_by ON public.ai_chat_sessions USING btree (created_by_user_id, last_activity_at DESC);


--
-- Name: idx_ai_chat_sessions_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_chat_sessions_project_id ON public.ai_chat_sessions USING btree (project_id, last_activity_at DESC);


--
-- Name: idx_ai_model_catalog_cache_provider; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_model_catalog_cache_provider ON public.ai_model_catalog_cache USING btree (provider_key, fetched_at DESC);


--
-- Name: idx_ai_model_evaluations_project_dependency; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_model_evaluations_project_dependency ON public.ai_model_evaluations USING btree (project_id, action_dependency_id, created_at DESC);


--
-- Name: idx_ai_patch_proposals_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_patch_proposals_project_id ON public.ai_patch_proposals USING btree (project_id, created_at DESC);


--
-- Name: idx_ai_prompt_snapshots_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_prompt_snapshots_project_id ON public.ai_prompt_snapshots USING btree (project_id, compiled_at DESC);


--
-- Name: idx_ai_session_context_frames_session; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_session_context_frames_session ON public.ai_session_context_frames USING btree (ai_chat_session_id, created_at DESC);


--
-- Name: idx_ai_session_context_frames_session_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_session_context_frames_session_created ON public.ai_session_context_frames USING btree (ai_chat_session_id, created_at DESC);


--
-- Name: idx_ai_tool_call_edits_tool_call; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_tool_call_edits_tool_call ON public.ai_tool_call_edits USING btree (ai_tool_call_id, created_at DESC);


--
-- Name: idx_ai_tool_calls_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_tool_calls_session_id ON public.ai_tool_calls USING btree (ai_chat_session_id, created_at DESC);


--
-- Name: idx_ai_tool_plan_steps_idempotency; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_tool_plan_steps_idempotency ON public.ai_tool_plan_steps USING btree (project_id, idempotency_key) WHERE (idempotency_key <> ''::text);


--
-- Name: idx_ai_tool_plan_steps_plan_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_tool_plan_steps_plan_status ON public.ai_tool_plan_steps USING btree (ai_tool_plan_id, status, created_at);


--
-- Name: idx_ai_tool_plans_project_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_tool_plans_project_status ON public.ai_tool_plans USING btree (project_id, status, created_at DESC);


--
-- Name: idx_ai_tool_plans_session_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_tool_plans_session_created ON public.ai_tool_plans USING btree (ai_chat_session_id, created_at DESC);


--
-- Name: idx_ai_usage_events_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_usage_events_project_id ON public.ai_usage_events USING btree (project_id, created_at DESC);


--
-- Name: idx_api_key_grants_lookup; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_api_key_grants_lookup ON public.api_key_grants USING btree (scope_kind, scope_id);


--
-- Name: idx_api_keys_user_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_api_keys_user_active ON public.api_keys USING btree (user_id, created_at DESC) WHERE (revoked_at IS NULL);


--
-- Name: idx_auth_events_user_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_auth_events_user_created_at ON public.auth_events USING btree (user_id, created_at DESC);


--
-- Name: idx_auth_identities_primary_user_unique_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_auth_identities_primary_user_unique_active ON public.auth_identities USING btree (user_id) WHERE ((is_primary IS TRUE) AND (deleted_at IS NULL));


--
-- Name: idx_auth_identities_provider_subject_unique_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_auth_identities_provider_subject_unique_active ON public.auth_identities USING btree (provider, provider_subject) WHERE (deleted_at IS NULL);


--
-- Name: idx_auth_identities_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_auth_identities_user_id ON public.auth_identities USING btree (user_id);


--
-- Name: idx_auth_identity_versions_auth_identity_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_auth_identity_versions_auth_identity_id ON public.auth_identity_versions USING btree (auth_identity_id);


--
-- Name: idx_auth_identity_versions_changed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_auth_identity_versions_changed_at ON public.auth_identity_versions USING btree (changed_at DESC);


--
-- Name: idx_automation_rules_project_enabled; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_automation_rules_project_enabled ON public.automation_rules USING btree (project_id, enabled);


--
-- Name: idx_bot_sessions_dependency_chat; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_bot_sessions_dependency_chat ON public.bot_sessions USING btree (action_dependency_id, relay_key, chat_id);


--
-- Name: idx_bot_sessions_project_updated; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_bot_sessions_project_updated ON public.bot_sessions USING btree (project_id, updated_at DESC);


--
-- Name: idx_client_contacts_client_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_client_contacts_client_id ON public.client_contacts USING btree (client_id);


--
-- Name: idx_client_projects_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_client_projects_project_id ON public.client_projects USING btree (project_id);


--
-- Name: idx_client_responsibles_client_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_client_responsibles_client_id ON public.client_responsibles USING btree (client_id);


--
-- Name: idx_client_responsibles_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_client_responsibles_company_id ON public.client_responsibles USING btree (company_id);


--
-- Name: idx_client_responsibles_legal_entity_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_client_responsibles_legal_entity_id ON public.client_responsibles USING btree (company_legal_entity_id);


--
-- Name: idx_client_services_client_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_client_services_client_id ON public.client_services USING btree (client_id);


--
-- Name: idx_client_services_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_client_services_company_id ON public.client_services USING btree (company_id);


--
-- Name: idx_client_services_project_due; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_client_services_project_due ON public.client_services USING btree (project_id, next_due_date) WHERE (status = 'active'::text);


--
-- Name: idx_clients_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_clients_company_id ON public.clients USING btree (company_id);


--
-- Name: idx_clients_company_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_clients_company_status ON public.clients USING btree (company_id, status);


--
-- Name: idx_clients_inn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_clients_inn ON public.clients USING btree (inn);


--
-- Name: idx_clients_search_vector; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_clients_search_vector ON public.clients USING gin (search_vector);


--
-- Name: idx_companies_owner_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_companies_owner_user_id ON public.companies USING btree (owner_user_id);


--
-- Name: idx_companies_slug_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_companies_slug_unique ON public.companies USING btree (slug) WHERE (slug IS NOT NULL);


--
-- Name: idx_company_legal_entities_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_company_legal_entities_company_id ON public.company_legal_entities USING btree (company_id);


--
-- Name: idx_company_legal_entities_inn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_company_legal_entities_inn ON public.company_legal_entities USING btree (inn);


--
-- Name: idx_company_membership_versions_changed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_company_membership_versions_changed_at ON public.company_membership_versions USING btree (changed_at DESC);


--
-- Name: idx_company_membership_versions_entity; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_company_membership_versions_entity ON public.company_membership_versions USING btree (company_id, user_id);


--
-- Name: idx_company_memberships_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_company_memberships_company_id ON public.company_memberships USING btree (company_id);


--
-- Name: idx_company_memberships_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_company_memberships_user_id ON public.company_memberships USING btree (user_id);


--
-- Name: idx_company_versions_changed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_company_versions_changed_at ON public.company_versions USING btree (changed_at DESC);


--
-- Name: idx_company_versions_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_company_versions_company_id ON public.company_versions USING btree (company_id);


--
-- Name: idx_desktop_auth_requests_device_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_desktop_auth_requests_device_code ON public.desktop_auth_requests USING btree (device_code);


--
-- Name: idx_desktop_auth_requests_user_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_desktop_auth_requests_user_code ON public.desktop_auth_requests USING btree (user_code);


--
-- Name: idx_desktop_devices_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_desktop_devices_project_id ON public.desktop_devices USING btree (project_id);


--
-- Name: idx_desktop_devices_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_desktop_devices_user_id ON public.desktop_devices USING btree (user_id);


--
-- Name: idx_desktop_messages_device_status_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_desktop_messages_device_status_created ON public.desktop_messages USING btree (desktop_device_id, status, created_at DESC);


--
-- Name: idx_desktop_messages_project_id_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_desktop_messages_project_id_created_at ON public.desktop_messages USING btree (project_id, created_at DESC);


--
-- Name: idx_device_credential_versions_changed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_device_credential_versions_changed_at ON public.device_credential_versions USING btree (changed_at DESC);


--
-- Name: idx_device_credential_versions_device_credential_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_device_credential_versions_device_credential_id ON public.device_credential_versions USING btree (device_credential_id);


--
-- Name: idx_device_credentials_fingerprint_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_device_credentials_fingerprint_hash ON public.device_credentials USING btree (device_fingerprint_hash);


--
-- Name: idx_device_credentials_user_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_device_credentials_user_active ON public.device_credentials USING btree (user_id, expires_at) WHERE (revoked_at IS NULL);


--
-- Name: idx_effect_ledger_run_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_effect_ledger_run_action ON public.effect_ledger USING btree (workflow_run_id, action_key, attempt_no);


--
-- Name: idx_effect_ledger_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_effect_ledger_status ON public.effect_ledger USING btree (status, updated_at);


--
-- Name: idx_email_templates_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_templates_project_id ON public.email_templates USING btree (project_id);


--
-- Name: idx_employee_profiles_company_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_employee_profiles_company_status ON public.employee_profiles USING btree (company_id, status);


--
-- Name: idx_integration_doc_chunks_search_vector; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_integration_doc_chunks_search_vector ON public.integration_doc_chunks USING gin (search_vector);


--
-- Name: idx_integration_docs_action_type_keys; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_integration_docs_action_type_keys ON public.integration_docs USING gin (action_type_keys);


--
-- Name: idx_integration_docs_dependency_type_keys; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_integration_docs_dependency_type_keys ON public.integration_docs USING gin (dependency_type_keys);


--
-- Name: idx_integration_docs_provider_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_integration_docs_provider_key ON public.integration_docs USING btree (provider_key);


--
-- Name: idx_integration_docs_search_vector; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_integration_docs_search_vector ON public.integration_docs USING gin (search_vector);


--
-- Name: idx_integration_docs_summary_trgm; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_integration_docs_summary_trgm ON public.integration_docs USING gin (summary public.gin_trgm_ops);


--
-- Name: idx_integration_docs_title_trgm; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_integration_docs_title_trgm ON public.integration_docs USING gin (title public.gin_trgm_ops);


--
-- Name: idx_message_templates_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_message_templates_project_id ON public.message_templates USING btree (project_id);


--
-- Name: idx_module_channel_bindings_lookup; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_module_channel_bindings_lookup ON public.module_channel_bindings USING btree (company_id, project_id, module_key, event_key) WHERE (enabled = true);


--
-- Name: idx_module_workflow_triggers_lookup; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_module_workflow_triggers_lookup ON public.module_workflow_triggers USING btree (company_id, project_id, module_key, event_key) WHERE (enabled = true);


--
-- Name: idx_notifications_user_unread; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user_unread ON public.notifications USING btree (user_id, created_at DESC) WHERE (read_at IS NULL);


--
-- Name: idx_outbox_events_aggregate; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_outbox_events_aggregate ON public.outbox_events USING btree (aggregate_type, aggregate_id, created_at) WHERE ((aggregate_type <> ''::text) AND (aggregate_id <> ''::text));


--
-- Name: idx_outbox_events_claim; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_outbox_events_claim ON public.outbox_events USING btree (status, available_at, created_at) WHERE (status = ANY (ARRAY['pending'::text, 'processing'::text]));


--
-- Name: idx_outbox_events_type_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_outbox_events_type_status ON public.outbox_events USING btree (event_type, status, created_at);


--
-- Name: idx_payment_action_documents_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_action_documents_action ON public.payment_action_documents USING btree (payment_action_id, created_at DESC);


--
-- Name: idx_payment_action_documents_project; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_action_documents_project ON public.payment_action_documents USING btree (project_id, created_at DESC);


--
-- Name: idx_payment_action_items_action_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_action_items_action_id ON public.payment_action_items USING btree (payment_action_id);


--
-- Name: idx_payment_actions_client_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_actions_client_id ON public.payment_actions USING btree (client_id);


--
-- Name: idx_payment_actions_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_actions_company_id ON public.payment_actions USING btree (company_id);


--
-- Name: idx_payment_actions_project_status_due; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_actions_project_status_due ON public.payment_actions USING btree (project_id, status, due_date);


--
-- Name: idx_payment_events_client_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_events_client_created ON public.payment_events USING btree (client_id, created_at DESC);


--
-- Name: idx_payment_events_project_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_events_project_created ON public.payment_events USING btree (project_id, created_at DESC);


--
-- Name: idx_payment_facts_action_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_facts_action_id ON public.payment_facts USING btree (payment_action_id);


--
-- Name: idx_payment_facts_project_paid_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_facts_project_paid_at ON public.payment_facts USING btree (project_id, paid_at DESC);


--
-- Name: idx_project_ai_employees_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_ai_employees_project_id ON public.project_ai_employees USING btree (project_id);


--
-- Name: idx_project_member_versions_changed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_member_versions_changed_at ON public.project_member_versions USING btree (changed_at DESC);


--
-- Name: idx_project_member_versions_entity; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_member_versions_entity ON public.project_member_versions USING btree (project_id, user_id);


--
-- Name: idx_project_members_client_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_members_client_id ON public.project_members USING btree (client_id);


--
-- Name: idx_project_members_project_role; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_members_project_role ON public.project_members USING btree (project_id, role);


--
-- Name: idx_project_members_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_members_user_id ON public.project_members USING btree (user_id);


--
-- Name: idx_project_versions_changed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_versions_changed_at ON public.project_versions USING btree (changed_at DESC);


--
-- Name: idx_project_versions_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_versions_project_id ON public.project_versions USING btree (project_id);


--
-- Name: idx_projects_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_company_id ON public.projects USING btree (company_id);


--
-- Name: idx_projects_hook_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_hook_key ON public.projects USING btree (hook_key);


--
-- Name: idx_provider_operation_effects_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_provider_operation_effects_status ON public.provider_operation_effects USING btree (provider, status, updated_at);


--
-- Name: idx_run_events_workflow_run_id_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_run_events_workflow_run_id_created_at ON public.run_events USING btree (workflow_run_id, created_at DESC);


--
-- Name: idx_secrets_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_secrets_company_id ON public.secrets USING btree (company_id);


--
-- Name: idx_secrets_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_secrets_project_id ON public.secrets USING btree (project_id);


--
-- Name: idx_sessions_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_expires_at ON public.sessions USING btree (expires_at);


--
-- Name: idx_sessions_user_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_user_active ON public.sessions USING btree (user_id, expires_at) WHERE (revoked_at IS NULL);


--
-- Name: idx_site_call_sessions_conversation_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_call_sessions_conversation_time ON public.site_call_sessions USING btree (conversation_id, created_at DESC) WHERE (conversation_id IS NOT NULL);


--
-- Name: idx_site_call_sessions_project_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_call_sessions_project_status ON public.site_call_sessions USING btree (project_id, status, created_at DESC);


--
-- Name: idx_site_call_sessions_visit_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_call_sessions_visit_status ON public.site_call_sessions USING btree (visit_session_id, status) WHERE (visit_session_id IS NOT NULL);


--
-- Name: idx_site_chat_attachments_call; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_chat_attachments_call ON public.site_chat_attachments USING btree (call_session_id, created_at);


--
-- Name: idx_site_chat_attachments_conversation; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_chat_attachments_conversation ON public.site_chat_attachments USING btree (conversation_id, created_at);


--
-- Name: idx_site_chat_attachments_message; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_chat_attachments_message ON public.site_chat_attachments USING btree (message_id, created_at);


--
-- Name: idx_site_chat_conversations_project_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_chat_conversations_project_status ON public.site_chat_conversations USING btree (project_id, status, updated_at DESC);


--
-- Name: idx_site_chat_conversations_visit_session; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_chat_conversations_visit_session ON public.site_chat_conversations USING btree (visit_session_id) WHERE (visit_session_id IS NOT NULL);


--
-- Name: idx_site_chat_messages_conversation_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_chat_messages_conversation_time ON public.site_chat_messages USING btree (conversation_id, created_at);


--
-- Name: idx_site_chat_widgets_project; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_chat_widgets_project ON public.site_chat_widgets USING btree (project_id, status, created_at DESC);


--
-- Name: idx_site_cobrowse_frames_session_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_cobrowse_frames_session_time ON public.site_cobrowse_frames USING btree (visit_session_id, occurred_at DESC);


--
-- Name: idx_site_visit_connections_project_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_visit_connections_project_status ON public.site_visit_connections USING btree (project_id, status, last_heartbeat_at DESC);


--
-- Name: idx_site_visit_connections_session_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_visit_connections_session_status ON public.site_visit_connections USING btree (visit_session_id, status, last_heartbeat_at DESC) WHERE (visit_session_id IS NOT NULL);


--
-- Name: idx_site_visit_sessions_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_visit_sessions_active ON public.site_visit_sessions USING btree (project_id, last_activity_at DESC) WHERE (status = ANY (ARRAY['active'::text, 'idle'::text]));


--
-- Name: idx_site_visit_sessions_project_duration; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_visit_sessions_project_duration ON public.site_visit_sessions USING btree (project_id, duration_seconds DESC, last_activity_at DESC);


--
-- Name: idx_site_visitor_events_session_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_visitor_events_session_time ON public.site_visitor_events USING btree (visit_session_id, occurred_at DESC);


--
-- Name: idx_site_visitors_project_last_seen; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_site_visitors_project_last_seen ON public.site_visitors USING btree (project_id, last_seen_at DESC);


--
-- Name: idx_task_activity_events_task_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_activity_events_task_created ON public.task_activity_events USING btree (task_id, created_at DESC);


--
-- Name: idx_task_attachments_project; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_attachments_project ON public.task_attachments USING btree (project_id, created_at DESC);


--
-- Name: idx_task_attachments_task; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_attachments_task ON public.task_attachments USING btree (task_id, created_at DESC);


--
-- Name: idx_task_boards_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_boards_project_id ON public.task_boards USING btree (project_id);


--
-- Name: idx_task_checklist_items_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_checklist_items_task_id ON public.task_checklist_items USING btree (task_id);


--
-- Name: idx_task_comments_task_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_comments_task_created ON public.task_comments USING btree (task_id, created_at);


--
-- Name: idx_task_custom_field_values_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_custom_field_values_task_id ON public.task_custom_field_values USING btree (task_id);


--
-- Name: idx_task_custom_fields_project_workflow; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_custom_fields_project_workflow ON public.task_custom_fields USING btree (project_id, task_workflow_id, "position");


--
-- Name: idx_task_process_memberships_role; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_process_memberships_role ON public.task_process_memberships USING btree (task_process_role_id);


--
-- Name: idx_task_process_memberships_user_project; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_process_memberships_user_project ON public.task_process_memberships USING btree (user_id, project_id) WHERE (status = 'active'::text);


--
-- Name: idx_task_process_roles_workflow_position; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_process_roles_workflow_position ON public.task_process_roles USING btree (task_workflow_id, "position");


--
-- Name: idx_task_status_history_task_changed; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_status_history_task_changed ON public.task_status_history USING btree (task_id, changed_at DESC);


--
-- Name: idx_task_views_project_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_views_project_user ON public.task_views USING btree (project_id, user_id);


--
-- Name: idx_task_workflow_statuses_workflow_position; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_workflow_statuses_workflow_position ON public.task_workflow_statuses USING btree (task_workflow_id, "position");


--
-- Name: idx_task_workflow_transitions_workflow; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_workflow_transitions_workflow ON public.task_workflow_transitions USING btree (task_workflow_id, from_status_id, to_status_id);


--
-- Name: idx_task_workflows_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_workflows_company_id ON public.task_workflows USING btree (company_id);


--
-- Name: idx_task_workflows_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_task_workflows_project_id ON public.task_workflows USING btree (project_id);


--
-- Name: idx_tasks_assignee_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tasks_assignee_user_id ON public.tasks USING btree (assignee_user_id);


--
-- Name: idx_tasks_client_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tasks_client_id ON public.tasks USING btree (client_id);


--
-- Name: idx_tasks_company_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tasks_company_status ON public.tasks USING btree (company_id, task_workflow_status_id);


--
-- Name: idx_tasks_due_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tasks_due_date ON public.tasks USING btree (due_date);


--
-- Name: idx_tasks_project_status_rank; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tasks_project_status_rank ON public.tasks USING btree (project_id, task_workflow_status_id, rank);


--
-- Name: idx_tasks_search_vector; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tasks_search_vector ON public.tasks USING gin (search_vector);


--
-- Name: idx_tasks_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tasks_updated_at ON public.tasks USING btree (updated_at DESC);


--
-- Name: idx_team_members_employee_profile_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_members_employee_profile_id ON public.team_members USING btree (employee_profile_id);


--
-- Name: idx_telegram_relay_configs_polling_due; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_telegram_relay_configs_polling_due ON public.telegram_relay_configs USING btree (polling_enabled, polling_locked_until) WHERE (status = 'active'::text);


--
-- Name: idx_telegram_relay_configs_relay_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_telegram_relay_configs_relay_key ON public.telegram_relay_configs USING btree (relay_key) WHERE (status = 'active'::text);


--
-- Name: idx_telegram_relay_events_dependency_received; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_telegram_relay_events_dependency_received ON public.telegram_relay_events USING btree (action_dependency_id, received_at DESC);


--
-- Name: idx_telegram_relay_events_project_received; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_telegram_relay_events_project_received ON public.telegram_relay_events USING btree (project_id, received_at DESC);


--
-- Name: idx_telegram_relay_forwards_event; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_telegram_relay_forwards_event ON public.telegram_relay_forwards USING btree (telegram_relay_event_id);


--
-- Name: idx_time_entries_client_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_entries_client_date ON public.time_entries USING btree (client_id, entry_date DESC);


--
-- Name: idx_time_entries_company_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_entries_company_date ON public.time_entries USING btree (company_id, entry_date DESC);


--
-- Name: idx_time_entries_project_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_entries_project_date ON public.time_entries USING btree (project_id, entry_date DESC);


--
-- Name: idx_time_entries_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_entries_task_id ON public.time_entries USING btree (task_id);


--
-- Name: idx_time_entries_user_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_entries_user_date ON public.time_entries USING btree (user_id, entry_date DESC);


--
-- Name: idx_timer_sessions_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_timer_sessions_task_id ON public.timer_sessions USING btree (task_id);


--
-- Name: idx_timer_sessions_user_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_timer_sessions_user_active ON public.timer_sessions USING btree (user_id, started_at DESC) WHERE (state = ANY (ARRAY['running'::text, 'paused'::text]));


--
-- Name: idx_user_versions_changed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_versions_changed_at ON public.user_versions USING btree (changed_at DESC);


--
-- Name: idx_user_versions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_versions_user_id ON public.user_versions USING btree (user_id);


--
-- Name: idx_users_current_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_current_company_id ON public.users USING btree (current_company_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_email_unique_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email_unique_active ON public.users USING btree (email) WHERE ((deleted_at IS NULL) AND (email IS NOT NULL));


--
-- Name: idx_webhook_endpoints_endpoint_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_webhook_endpoints_endpoint_key ON public.webhook_endpoints USING btree (endpoint_key);


--
-- Name: idx_webhook_endpoints_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_webhook_endpoints_project_id ON public.webhook_endpoints USING btree (project_id);


--
-- Name: idx_webhook_events_endpoint_body_sha256; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_webhook_events_endpoint_body_sha256 ON public.webhook_events USING btree (webhook_endpoint_id, body_sha256, received_at DESC) WHERE (body_sha256 IS NOT NULL);


--
-- Name: idx_webhook_events_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_webhook_events_status ON public.webhook_events USING btree (status);


--
-- Name: idx_webhook_events_webhook_endpoint_id_received_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_webhook_events_webhook_endpoint_id_received_at ON public.webhook_events USING btree (webhook_endpoint_id, received_at DESC);


--
-- Name: idx_workflow_action_dead_letters_owner_class; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_action_dead_letters_owner_class ON public.workflow_action_dead_letters USING btree (owner, retry_class, created_at DESC);


--
-- Name: idx_workflow_action_dead_letters_replayable; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_action_dead_letters_replayable ON public.workflow_action_dead_letters USING btree (replayable, created_at DESC) WHERE (replayable = true);


--
-- Name: idx_workflow_client_payment_due_state_enabled; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_client_payment_due_state_enabled ON public.workflow_client_payment_due_state USING btree (project_id, updated_at) WHERE (enabled = true);


--
-- Name: idx_workflow_runs_cancelled_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_runs_cancelled_at ON public.workflow_runs USING btree (cancelled_at DESC, cancel_reason) WHERE (status = 'cancelled'::text);


--
-- Name: idx_workflow_runs_claim_lease; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_runs_claim_lease ON public.workflow_runs USING btree (status, run_lease_expires_at, workflow_run_id) WHERE (status = ANY (ARRAY['pending'::text, 'running'::text]));


--
-- Name: idx_workflow_runs_claim_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_runs_claim_owner ON public.workflow_runs USING btree (run_claim_owner, run_fencing_token) WHERE (run_claim_owner <> ''::text);


--
-- Name: idx_workflow_runs_project_id_started_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_runs_project_id_started_at ON public.workflow_runs USING btree (project_id, started_at DESC NULLS LAST);


--
-- Name: idx_workflow_runs_run_request_source; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_runs_run_request_source ON public.workflow_runs USING btree (source_type, run_request_version, workflow_run_id);


--
-- Name: idx_workflow_runs_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_runs_status ON public.workflow_runs USING btree (status);


--
-- Name: idx_workflow_runs_waiting_deadline; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_runs_waiting_deadline ON public.workflow_runs USING btree (waiting_deadline_at, workflow_run_id) WHERE ((status = 'waiting'::text) AND (waiting_deadline_at IS NOT NULL));


--
-- Name: idx_workflow_runs_workflow_id_started_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_runs_workflow_id_started_at ON public.workflow_runs USING btree (workflow_id, started_at DESC NULLS LAST);


--
-- Name: idx_workflow_schedule_state_due; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_schedule_state_due ON public.workflow_schedule_state USING btree (next_run_at) WHERE (enabled = true);


--
-- Name: idx_workflow_versions_workflow_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflow_versions_workflow_id ON public.workflow_versions USING btree (workflow_id);


--
-- Name: idx_workflows_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflows_project_id ON public.workflows USING btree (project_id);


--
-- Name: uidx_ai_agent_events_run_seq; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_ai_agent_events_run_seq ON public.ai_agent_events USING btree (ai_agent_run_id, seq) WHERE (ai_agent_run_id IS NOT NULL);


--
-- Name: uidx_ai_agent_runs_project_idempotency; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_ai_agent_runs_project_idempotency ON public.ai_agent_runs USING btree (project_id, idempotency_key);


--
-- Name: uidx_ai_chat_cards_session_card; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_ai_chat_cards_session_card ON public.ai_chat_cards USING btree (ai_chat_session_id, card_id);


--
-- Name: uidx_ai_tool_calls_idempotency_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_ai_tool_calls_idempotency_key ON public.ai_tool_calls USING btree (project_id, idempotency_key) WHERE (idempotency_key <> ''::text);


--
-- Name: uidx_ai_tool_idempotency_project_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_ai_tool_idempotency_project_key ON public.ai_tool_idempotency USING btree (project_id, idempotency_key);


--
-- Name: uidx_api_key_grants_scope; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_api_key_grants_scope ON public.api_key_grants USING btree (api_key_id, scope_kind, COALESCE(scope_id, '00000000-0000-0000-0000-000000000000'::uuid));


--
-- Name: uidx_client_contacts_one_primary; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_client_contacts_one_primary ON public.client_contacts USING btree (client_id) WHERE ((is_primary = true) AND (status = 'active'::text));


--
-- Name: uidx_client_responsibles_one_primary; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_client_responsibles_one_primary ON public.client_responsibles USING btree (client_id) WHERE ((is_primary = true) AND (status = 'active'::text));


--
-- Name: uidx_client_service_types_global_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_client_service_types_global_key ON public.client_service_types USING btree (service_type_key) WHERE (project_id IS NULL);


--
-- Name: uidx_client_service_types_project_active_display_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_client_service_types_project_active_display_name ON public.client_service_types USING btree (project_id, lower(btrim(display_name))) WHERE ((project_id IS NOT NULL) AND (status = 'active'::text));


--
-- Name: uidx_client_service_types_project_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_client_service_types_project_key ON public.client_service_types USING btree (project_id, service_type_key) WHERE (project_id IS NOT NULL);


--
-- Name: uidx_clients_company_inn; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_clients_company_inn ON public.clients USING btree (company_id, inn) WHERE ((inn IS NOT NULL) AND (inn <> ''::text) AND (status <> 'deleted'::text));


--
-- Name: uidx_company_legal_entities_company_inn; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_company_legal_entities_company_inn ON public.company_legal_entities USING btree (company_id, inn) WHERE ((inn IS NOT NULL) AND (inn <> ''::text) AND (status <> 'deleted'::text));


--
-- Name: uidx_employee_profiles_company_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_employee_profiles_company_user ON public.employee_profiles USING btree (company_id, user_id);


--
-- Name: uidx_module_channel_bindings_company_channel; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_module_channel_bindings_company_channel ON public.module_channel_bindings USING btree (company_id, module_key, event_key, channel_key) WHERE (project_id IS NULL);


--
-- Name: uidx_module_channel_bindings_project_channel; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_module_channel_bindings_project_channel ON public.module_channel_bindings USING btree (company_id, project_id, module_key, event_key, channel_key) WHERE (project_id IS NOT NULL);


--
-- Name: uidx_module_workflow_triggers_company_event_workflow; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_module_workflow_triggers_company_event_workflow ON public.module_workflow_triggers USING btree (company_id, module_key, event_key, workflow_id) WHERE (project_id IS NULL);


--
-- Name: uidx_module_workflow_triggers_project_event_workflow; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_module_workflow_triggers_project_event_workflow ON public.module_workflow_triggers USING btree (company_id, project_id, module_key, event_key, workflow_id) WHERE (project_id IS NOT NULL);


--
-- Name: uidx_payment_action_items_service; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_payment_action_items_service ON public.payment_action_items USING btree (payment_action_id, client_service_id);


--
-- Name: uidx_payment_actions_open_period; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_payment_actions_open_period ON public.payment_actions USING btree (client_id, project_id, period_start) WHERE (status = ANY (ARRAY['pending'::text, 'sent'::text, 'snoozed'::text]));


--
-- Name: uidx_payment_events_source_idempotency; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_payment_events_source_idempotency ON public.payment_events USING btree (event_source, idempotency_key);


--
-- Name: uidx_projects_company_project_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_projects_company_project_key ON public.projects USING btree (company_id, project_key) WHERE (project_key IS NOT NULL);


--
-- Name: uidx_secrets_company_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_secrets_company_name ON public.secrets USING btree (company_id, name) WHERE ((scope_kind = 'company'::text) AND (company_id IS NOT NULL));


--
-- Name: uidx_secrets_project_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_secrets_project_name ON public.secrets USING btree (project_id, name) WHERE (project_id IS NOT NULL);


--
-- Name: uidx_site_cobrowse_frames_session_sequence; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_site_cobrowse_frames_session_sequence ON public.site_cobrowse_frames USING btree (visit_session_id, sequence);


--
-- Name: uidx_site_visitors_widget_anon; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_site_visitors_widget_anon ON public.site_visitors USING btree (chat_widget_id, anonymous_id);


--
-- Name: uidx_task_boards_project_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_task_boards_project_name ON public.task_boards USING btree (project_id, lower(name)) WHERE (project_id IS NOT NULL);


--
-- Name: uidx_task_custom_field_values_field_task; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_task_custom_field_values_field_task ON public.task_custom_field_values USING btree (task_custom_field_id, task_id);


--
-- Name: uidx_task_process_memberships_workflow_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_task_process_memberships_workflow_user ON public.task_process_memberships USING btree (task_workflow_id, user_id);


--
-- Name: uidx_task_process_roles_workflow_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_task_process_roles_workflow_key ON public.task_process_roles USING btree (task_workflow_id, role_key);


--
-- Name: uidx_tasks_company_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_tasks_company_key ON public.tasks USING btree (company_id, task_key);


--
-- Name: uidx_telegram_relay_events_dependency_update; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_telegram_relay_events_dependency_update ON public.telegram_relay_events USING btree (action_dependency_id, update_id) WHERE (update_id IS NOT NULL);


--
-- Name: uidx_webhook_events_endpoint_event_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_webhook_events_endpoint_event_id ON public.webhook_events USING btree (webhook_endpoint_id, event_id) WHERE (event_id IS NOT NULL);


--
-- Name: uidx_webhook_events_endpoint_idempotency; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_webhook_events_endpoint_idempotency ON public.webhook_events USING btree (webhook_endpoint_id, idempotency_key) WHERE (idempotency_key IS NOT NULL);


--
-- Name: uidx_workflow_runs_run_request_idempotency_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_workflow_runs_run_request_idempotency_key ON public.workflow_runs USING btree (run_request_idempotency_key) WHERE (run_request_idempotency_key <> ''::text);


--
-- Name: action_dependencies trg_action_dependencies_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_action_dependencies_set_updated_at BEFORE UPDATE ON public.action_dependencies FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: action_dependency_types trg_action_dependency_types_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_action_dependency_types_set_updated_at BEFORE UPDATE ON public.action_dependency_types FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: action_types trg_action_types_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_action_types_set_updated_at BEFORE UPDATE ON public.action_types FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: ai_chat_sessions trg_ai_chat_sessions_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_ai_chat_sessions_set_updated_at BEFORE UPDATE ON public.ai_chat_sessions FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: api_keys trg_api_keys_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_api_keys_set_updated_at BEFORE UPDATE ON public.api_keys FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: auth_identities trg_auth_identities_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_auth_identities_set_updated_at BEFORE UPDATE ON public.auth_identities FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: auth_identities trg_auth_identity_versions_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_auth_identity_versions_delete BEFORE DELETE ON public.auth_identities FOR EACH ROW EXECUTE FUNCTION public.auth_identity_versions_on_delete();


--
-- Name: auth_identities trg_auth_identity_versions_write; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_auth_identity_versions_write AFTER INSERT OR UPDATE ON public.auth_identities FOR EACH ROW EXECUTE FUNCTION public.auth_identity_versions_on_write();


--
-- Name: client_contacts trg_client_contacts_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_client_contacts_set_updated_at BEFORE UPDATE ON public.client_contacts FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: client_responsibles trg_client_responsibles_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_client_responsibles_set_updated_at BEFORE UPDATE ON public.client_responsibles FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: client_service_types trg_client_service_types_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_client_service_types_set_updated_at BEFORE UPDATE ON public.client_service_types FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: client_services trg_client_services_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_client_services_set_updated_at BEFORE UPDATE ON public.client_services FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: clients trg_clients_set_search_vector; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_clients_set_search_vector BEFORE INSERT OR UPDATE OF name, short_name, full_name, inn, kpp, ogrn, ogrnip, comment ON public.clients FOR EACH ROW EXECUTE FUNCTION public.set_clients_search_vector();


--
-- Name: clients trg_clients_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_clients_set_updated_at BEFORE UPDATE ON public.clients FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: companies trg_companies_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_companies_set_updated_at BEFORE UPDATE ON public.companies FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: company_legal_entities trg_company_legal_entities_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_company_legal_entities_set_updated_at BEFORE UPDATE ON public.company_legal_entities FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: company_memberships trg_company_membership_versions_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_company_membership_versions_delete BEFORE DELETE ON public.company_memberships FOR EACH ROW EXECUTE FUNCTION public.company_membership_versions_on_delete();


--
-- Name: company_memberships trg_company_membership_versions_write; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_company_membership_versions_write AFTER INSERT OR UPDATE ON public.company_memberships FOR EACH ROW EXECUTE FUNCTION public.company_membership_versions_on_write();


--
-- Name: companies trg_company_versions_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_company_versions_delete BEFORE DELETE ON public.companies FOR EACH ROW EXECUTE FUNCTION public.company_versions_on_delete();


--
-- Name: companies trg_company_versions_write; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_company_versions_write AFTER INSERT OR UPDATE ON public.companies FOR EACH ROW EXECUTE FUNCTION public.company_versions_on_write();


--
-- Name: device_credentials trg_device_credential_versions_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_device_credential_versions_delete BEFORE DELETE ON public.device_credentials FOR EACH ROW EXECUTE FUNCTION public.device_credential_versions_on_delete();


--
-- Name: device_credentials trg_device_credential_versions_write; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_device_credential_versions_write AFTER INSERT OR UPDATE ON public.device_credentials FOR EACH ROW EXECUTE FUNCTION public.device_credential_versions_on_write();


--
-- Name: device_credentials trg_device_credentials_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_device_credentials_set_updated_at BEFORE UPDATE ON public.device_credentials FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: email_templates trg_email_templates_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_email_templates_set_updated_at BEFORE UPDATE ON public.email_templates FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: employee_profiles trg_employee_profiles_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_employee_profiles_set_updated_at BEFORE UPDATE ON public.employee_profiles FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: integration_doc_chunks trg_integration_doc_chunks_set_search_vector; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_integration_doc_chunks_set_search_vector BEFORE INSERT OR UPDATE ON public.integration_doc_chunks FOR EACH ROW EXECUTE FUNCTION public.integration_doc_chunks_set_search_vector();


--
-- Name: integration_docs trg_integration_docs_set_search_vector; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_integration_docs_set_search_vector BEFORE INSERT OR UPDATE ON public.integration_docs FOR EACH ROW EXECUTE FUNCTION public.integration_docs_set_search_vector();


--
-- Name: message_templates trg_message_templates_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_message_templates_set_updated_at BEFORE UPDATE ON public.message_templates FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: module_channel_bindings trg_module_channel_bindings_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_module_channel_bindings_set_updated_at BEFORE UPDATE ON public.module_channel_bindings FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: module_workflow_triggers trg_module_workflow_triggers_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_module_workflow_triggers_set_updated_at BEFORE UPDATE ON public.module_workflow_triggers FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: payment_actions trg_payment_actions_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_payment_actions_set_updated_at BEFORE UPDATE ON public.payment_actions FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: project_ai_employees trg_project_ai_employees_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_project_ai_employees_set_updated_at BEFORE UPDATE ON public.project_ai_employees FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: project_members trg_project_member_versions_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_project_member_versions_delete BEFORE DELETE ON public.project_members FOR EACH ROW EXECUTE FUNCTION public.project_member_versions_on_delete();


--
-- Name: project_members trg_project_member_versions_write; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_project_member_versions_write AFTER INSERT OR UPDATE ON public.project_members FOR EACH ROW EXECUTE FUNCTION public.project_member_versions_on_write();


--
-- Name: projects trg_project_versions_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_project_versions_delete BEFORE DELETE ON public.projects FOR EACH ROW EXECUTE FUNCTION public.project_versions_on_delete();


--
-- Name: projects trg_project_versions_write; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_project_versions_write AFTER INSERT OR UPDATE ON public.projects FOR EACH ROW EXECUTE FUNCTION public.project_versions_on_write();


--
-- Name: projects trg_projects_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_projects_set_updated_at BEFORE UPDATE ON public.projects FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: secrets trg_secrets_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_secrets_set_updated_at BEFORE UPDATE ON public.secrets FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: sessions trg_sessions_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_sessions_set_updated_at BEFORE UPDATE ON public.sessions FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: site_call_sessions trg_site_call_sessions_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_site_call_sessions_set_updated_at BEFORE UPDATE ON public.site_call_sessions FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: site_chat_conversations trg_site_chat_conversations_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_site_chat_conversations_set_updated_at BEFORE UPDATE ON public.site_chat_conversations FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: site_chat_widgets trg_site_chat_widgets_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_site_chat_widgets_set_updated_at BEFORE UPDATE ON public.site_chat_widgets FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: site_visit_connections trg_site_visit_connections_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_site_visit_connections_set_updated_at BEFORE UPDATE ON public.site_visit_connections FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: site_visit_sessions trg_site_visit_sessions_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_site_visit_sessions_set_updated_at BEFORE UPDATE ON public.site_visit_sessions FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: site_visitors trg_site_visitors_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_site_visitors_set_updated_at BEFORE UPDATE ON public.site_visitors FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: task_boards trg_task_boards_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_task_boards_set_updated_at BEFORE UPDATE ON public.task_boards FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: task_comments trg_task_comments_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_task_comments_set_updated_at BEFORE UPDATE ON public.task_comments FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: task_custom_fields trg_task_custom_fields_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_task_custom_fields_set_updated_at BEFORE UPDATE ON public.task_custom_fields FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: task_process_memberships trg_task_process_memberships_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_task_process_memberships_set_updated_at BEFORE UPDATE ON public.task_process_memberships FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: task_process_roles trg_task_process_roles_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_task_process_roles_set_updated_at BEFORE UPDATE ON public.task_process_roles FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: task_workflow_statuses trg_task_workflow_statuses_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_task_workflow_statuses_set_updated_at BEFORE UPDATE ON public.task_workflow_statuses FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: task_workflows trg_task_workflows_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_task_workflows_set_updated_at BEFORE UPDATE ON public.task_workflows FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: tasks trg_tasks_set_search_vector; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_tasks_set_search_vector BEFORE INSERT OR UPDATE OF task_key, title, description, labels ON public.tasks FOR EACH ROW EXECUTE FUNCTION public.set_tasks_search_vector();


--
-- Name: tasks trg_tasks_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_tasks_set_updated_at BEFORE UPDATE ON public.tasks FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: teams trg_teams_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_teams_set_updated_at BEFORE UPDATE ON public.teams FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: time_entries trg_time_entries_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_time_entries_set_updated_at BEFORE UPDATE ON public.time_entries FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: timer_sessions trg_timer_sessions_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_timer_sessions_set_updated_at BEFORE UPDATE ON public.timer_sessions FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: users trg_user_versions_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_user_versions_delete BEFORE DELETE ON public.users FOR EACH ROW EXECUTE FUNCTION public.user_versions_on_delete();


--
-- Name: users trg_user_versions_write; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_user_versions_write AFTER INSERT OR UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.user_versions_on_write();


--
-- Name: user_workspace_contexts trg_user_workspace_contexts_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_user_workspace_contexts_set_updated_at BEFORE UPDATE ON public.user_workspace_contexts FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: users trg_users_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_users_set_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: webhook_endpoints trg_webhook_endpoints_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_webhook_endpoints_set_updated_at BEFORE UPDATE ON public.webhook_endpoints FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: workflow_runs trg_workflow_runs_enrich_payment_current_notification; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_workflow_runs_enrich_payment_current_notification BEFORE INSERT ON public.workflow_runs FOR EACH ROW EXECUTE FUNCTION public.enrich_payment_current_notification_input();


--
-- Name: workflows trg_workflows_set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_workflows_set_updated_at BEFORE UPDATE ON public.workflows FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_now();


--
-- Name: action_attempts action_attempts_action_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_attempts
    ADD CONSTRAINT action_attempts_action_run_id_fkey FOREIGN KEY (action_run_id) REFERENCES public.action_runs(action_run_id) ON DELETE CASCADE;


--
-- Name: action_dependencies action_dependencies_action_dependency_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_dependencies
    ADD CONSTRAINT action_dependencies_action_dependency_type_id_fkey FOREIGN KEY (action_dependency_type_id) REFERENCES public.action_dependency_types(action_dependency_type_id) ON DELETE RESTRICT;


--
-- Name: action_dependencies action_dependencies_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_dependencies
    ADD CONSTRAINT action_dependencies_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: action_dependencies action_dependencies_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_dependencies
    ADD CONSTRAINT action_dependencies_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: action_dependencies action_dependencies_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_dependencies
    ADD CONSTRAINT action_dependencies_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: action_dependencies action_dependencies_secret_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_dependencies
    ADD CONSTRAINT action_dependencies_secret_id_fkey FOREIGN KEY (secret_id) REFERENCES public.secrets(secret_id);


--
-- Name: action_runs action_runs_action_dependency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_runs
    ADD CONSTRAINT action_runs_action_dependency_id_fkey FOREIGN KEY (action_dependency_id) REFERENCES public.action_dependencies(action_dependency_id) ON DELETE SET NULL;


--
-- Name: action_runs action_runs_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_runs
    ADD CONSTRAINT action_runs_workflow_run_id_fkey FOREIGN KEY (workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE CASCADE;


--
-- Name: ai_agent_audit_events ai_agent_audit_events_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_audit_events
    ADD CONSTRAINT ai_agent_audit_events_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: ai_agent_audit_events ai_agent_audit_events_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_audit_events
    ADD CONSTRAINT ai_agent_audit_events_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_agent_audit_events ai_agent_audit_events_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_audit_events
    ADD CONSTRAINT ai_agent_audit_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: ai_agent_events ai_agent_events_ai_agent_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_events
    ADD CONSTRAINT ai_agent_events_ai_agent_run_id_fkey FOREIGN KEY (ai_agent_run_id) REFERENCES public.ai_agent_runs(ai_agent_run_id) ON DELETE CASCADE;


--
-- Name: ai_agent_events ai_agent_events_ai_chat_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_events
    ADD CONSTRAINT ai_agent_events_ai_chat_session_id_fkey FOREIGN KEY (ai_chat_session_id) REFERENCES public.ai_chat_sessions(ai_chat_session_id) ON DELETE CASCADE;


--
-- Name: ai_agent_events ai_agent_events_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_events
    ADD CONSTRAINT ai_agent_events_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_agent_runs ai_agent_runs_ai_chat_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_runs
    ADD CONSTRAINT ai_agent_runs_ai_chat_session_id_fkey FOREIGN KEY (ai_chat_session_id) REFERENCES public.ai_chat_sessions(ai_chat_session_id) ON DELETE CASCADE;


--
-- Name: ai_agent_runs ai_agent_runs_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_runs
    ADD CONSTRAINT ai_agent_runs_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: ai_agent_runs ai_agent_runs_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_runs
    ADD CONSTRAINT ai_agent_runs_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_agent_runs ai_agent_runs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_agent_runs
    ADD CONSTRAINT ai_agent_runs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: ai_channel_bindings ai_channel_bindings_action_dependency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_channel_bindings
    ADD CONSTRAINT ai_channel_bindings_action_dependency_id_fkey FOREIGN KEY (action_dependency_id) REFERENCES public.action_dependencies(action_dependency_id) ON DELETE SET NULL;


--
-- Name: ai_channel_bindings ai_channel_bindings_project_ai_employee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_channel_bindings
    ADD CONSTRAINT ai_channel_bindings_project_ai_employee_id_fkey FOREIGN KEY (project_ai_employee_id) REFERENCES public.project_ai_employees(project_ai_employee_id) ON DELETE CASCADE;


--
-- Name: ai_channel_bindings ai_channel_bindings_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_channel_bindings
    ADD CONSTRAINT ai_channel_bindings_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_chat_cards ai_chat_cards_ai_chat_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_cards
    ADD CONSTRAINT ai_chat_cards_ai_chat_session_id_fkey FOREIGN KEY (ai_chat_session_id) REFERENCES public.ai_chat_sessions(ai_chat_session_id) ON DELETE CASCADE;


--
-- Name: ai_chat_cards ai_chat_cards_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_cards
    ADD CONSTRAINT ai_chat_cards_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_chat_messages ai_chat_messages_ai_chat_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_messages
    ADD CONSTRAINT ai_chat_messages_ai_chat_session_id_fkey FOREIGN KEY (ai_chat_session_id) REFERENCES public.ai_chat_sessions(ai_chat_session_id) ON DELETE CASCADE;


--
-- Name: ai_chat_sessions ai_chat_sessions_action_dependency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_sessions
    ADD CONSTRAINT ai_chat_sessions_action_dependency_id_fkey FOREIGN KEY (action_dependency_id) REFERENCES public.action_dependencies(action_dependency_id) ON DELETE SET NULL;


--
-- Name: ai_chat_sessions ai_chat_sessions_active_patch_proposal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_sessions
    ADD CONSTRAINT ai_chat_sessions_active_patch_proposal_id_fkey FOREIGN KEY (active_patch_proposal_id) REFERENCES public.ai_patch_proposals(ai_patch_proposal_id) ON DELETE SET NULL;


--
-- Name: ai_chat_sessions ai_chat_sessions_active_tool_call_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_sessions
    ADD CONSTRAINT ai_chat_sessions_active_tool_call_id_fkey FOREIGN KEY (active_tool_call_id) REFERENCES public.ai_tool_calls(ai_tool_call_id) ON DELETE SET NULL;


--
-- Name: ai_chat_sessions ai_chat_sessions_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_sessions
    ADD CONSTRAINT ai_chat_sessions_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE SET NULL;


--
-- Name: ai_chat_sessions ai_chat_sessions_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_sessions
    ADD CONSTRAINT ai_chat_sessions_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: ai_chat_sessions ai_chat_sessions_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_sessions
    ADD CONSTRAINT ai_chat_sessions_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: ai_chat_sessions ai_chat_sessions_project_ai_employee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_sessions
    ADD CONSTRAINT ai_chat_sessions_project_ai_employee_id_fkey FOREIGN KEY (project_ai_employee_id) REFERENCES public.project_ai_employees(project_ai_employee_id) ON DELETE CASCADE;


--
-- Name: ai_chat_sessions ai_chat_sessions_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_sessions
    ADD CONSTRAINT ai_chat_sessions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_chat_sessions ai_chat_sessions_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_sessions
    ADD CONSTRAINT ai_chat_sessions_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(task_id) ON DELETE SET NULL;


--
-- Name: ai_chat_sessions ai_chat_sessions_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_sessions
    ADD CONSTRAINT ai_chat_sessions_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES public.workflows(workflow_id) ON DELETE SET NULL;


--
-- Name: ai_chat_sessions ai_chat_sessions_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_chat_sessions
    ADD CONSTRAINT ai_chat_sessions_workflow_run_id_fkey FOREIGN KEY (workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE SET NULL;


--
-- Name: ai_model_catalog_cache ai_model_catalog_cache_action_dependency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_model_catalog_cache
    ADD CONSTRAINT ai_model_catalog_cache_action_dependency_id_fkey FOREIGN KEY (action_dependency_id) REFERENCES public.action_dependencies(action_dependency_id) ON DELETE CASCADE;


--
-- Name: ai_model_evaluations ai_model_evaluations_action_dependency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_model_evaluations
    ADD CONSTRAINT ai_model_evaluations_action_dependency_id_fkey FOREIGN KEY (action_dependency_id) REFERENCES public.action_dependencies(action_dependency_id) ON DELETE CASCADE;


--
-- Name: ai_model_evaluations ai_model_evaluations_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_model_evaluations
    ADD CONSTRAINT ai_model_evaluations_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_patch_proposals ai_patch_proposals_ai_chat_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_patch_proposals
    ADD CONSTRAINT ai_patch_proposals_ai_chat_session_id_fkey FOREIGN KEY (ai_chat_session_id) REFERENCES public.ai_chat_sessions(ai_chat_session_id) ON DELETE CASCADE;


--
-- Name: ai_patch_proposals ai_patch_proposals_applied_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_patch_proposals
    ADD CONSTRAINT ai_patch_proposals_applied_by_user_id_fkey FOREIGN KEY (applied_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: ai_patch_proposals ai_patch_proposals_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_patch_proposals
    ADD CONSTRAINT ai_patch_proposals_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_patch_proposals ai_patch_proposals_reverted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_patch_proposals
    ADD CONSTRAINT ai_patch_proposals_reverted_by_user_id_fkey FOREIGN KEY (reverted_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: ai_prompt_snapshots ai_prompt_snapshots_ai_chat_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_prompt_snapshots
    ADD CONSTRAINT ai_prompt_snapshots_ai_chat_session_id_fkey FOREIGN KEY (ai_chat_session_id) REFERENCES public.ai_chat_sessions(ai_chat_session_id) ON DELETE CASCADE;


--
-- Name: ai_prompt_snapshots ai_prompt_snapshots_ai_tool_call_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_prompt_snapshots
    ADD CONSTRAINT ai_prompt_snapshots_ai_tool_call_id_fkey FOREIGN KEY (ai_tool_call_id) REFERENCES public.ai_tool_calls(ai_tool_call_id) ON DELETE SET NULL;


--
-- Name: ai_prompt_snapshots ai_prompt_snapshots_copied_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_prompt_snapshots
    ADD CONSTRAINT ai_prompt_snapshots_copied_by_user_id_fkey FOREIGN KEY (copied_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: ai_prompt_snapshots ai_prompt_snapshots_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_prompt_snapshots
    ADD CONSTRAINT ai_prompt_snapshots_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_session_context_frames ai_session_context_frames_ai_chat_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_session_context_frames
    ADD CONSTRAINT ai_session_context_frames_ai_chat_session_id_fkey FOREIGN KEY (ai_chat_session_id) REFERENCES public.ai_chat_sessions(ai_chat_session_id) ON DELETE CASCADE;


--
-- Name: ai_session_context_frames ai_session_context_frames_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_session_context_frames
    ADD CONSTRAINT ai_session_context_frames_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_tool_call_edits ai_tool_call_edits_ai_chat_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_call_edits
    ADD CONSTRAINT ai_tool_call_edits_ai_chat_session_id_fkey FOREIGN KEY (ai_chat_session_id) REFERENCES public.ai_chat_sessions(ai_chat_session_id) ON DELETE CASCADE;


--
-- Name: ai_tool_call_edits ai_tool_call_edits_ai_tool_call_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_call_edits
    ADD CONSTRAINT ai_tool_call_edits_ai_tool_call_id_fkey FOREIGN KEY (ai_tool_call_id) REFERENCES public.ai_tool_calls(ai_tool_call_id) ON DELETE CASCADE;


--
-- Name: ai_tool_call_edits ai_tool_call_edits_edited_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_call_edits
    ADD CONSTRAINT ai_tool_call_edits_edited_by_user_id_fkey FOREIGN KEY (edited_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: ai_tool_call_edits ai_tool_call_edits_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_call_edits
    ADD CONSTRAINT ai_tool_call_edits_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_tool_calls ai_tool_calls_ai_chat_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_calls
    ADD CONSTRAINT ai_tool_calls_ai_chat_session_id_fkey FOREIGN KEY (ai_chat_session_id) REFERENCES public.ai_chat_sessions(ai_chat_session_id) ON DELETE CASCADE;


--
-- Name: ai_tool_calls ai_tool_calls_approved_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_calls
    ADD CONSTRAINT ai_tool_calls_approved_by_user_id_fkey FOREIGN KEY (approved_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: ai_tool_calls ai_tool_calls_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_calls
    ADD CONSTRAINT ai_tool_calls_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_tool_calls ai_tool_calls_rejected_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_calls
    ADD CONSTRAINT ai_tool_calls_rejected_by_user_id_fkey FOREIGN KEY (rejected_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: ai_tool_idempotency ai_tool_idempotency_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_idempotency
    ADD CONSTRAINT ai_tool_idempotency_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_tool_plan_steps ai_tool_plan_steps_ai_tool_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_plan_steps
    ADD CONSTRAINT ai_tool_plan_steps_ai_tool_plan_id_fkey FOREIGN KEY (ai_tool_plan_id) REFERENCES public.ai_tool_plans(ai_tool_plan_id) ON DELETE CASCADE;


--
-- Name: ai_tool_plan_steps ai_tool_plan_steps_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_plan_steps
    ADD CONSTRAINT ai_tool_plan_steps_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_tool_plans ai_tool_plans_ai_chat_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_plans
    ADD CONSTRAINT ai_tool_plans_ai_chat_session_id_fkey FOREIGN KEY (ai_chat_session_id) REFERENCES public.ai_chat_sessions(ai_chat_session_id) ON DELETE CASCADE;


--
-- Name: ai_tool_plans ai_tool_plans_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_plans
    ADD CONSTRAINT ai_tool_plans_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: ai_tool_plans ai_tool_plans_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_plans
    ADD CONSTRAINT ai_tool_plans_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_tool_plans ai_tool_plans_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_tool_plans
    ADD CONSTRAINT ai_tool_plans_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: ai_usage_events ai_usage_events_action_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_usage_events
    ADD CONSTRAINT ai_usage_events_action_run_id_fkey FOREIGN KEY (action_run_id) REFERENCES public.action_runs(action_run_id) ON DELETE SET NULL;


--
-- Name: ai_usage_events ai_usage_events_ai_chat_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_usage_events
    ADD CONSTRAINT ai_usage_events_ai_chat_session_id_fkey FOREIGN KEY (ai_chat_session_id) REFERENCES public.ai_chat_sessions(ai_chat_session_id) ON DELETE SET NULL;


--
-- Name: ai_usage_events ai_usage_events_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_usage_events
    ADD CONSTRAINT ai_usage_events_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: ai_usage_events ai_usage_events_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_usage_events
    ADD CONSTRAINT ai_usage_events_workflow_run_id_fkey FOREIGN KEY (workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE SET NULL;


--
-- Name: api_key_grants api_key_grants_api_key_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_key_grants
    ADD CONSTRAINT api_key_grants_api_key_id_fkey FOREIGN KEY (api_key_id) REFERENCES public.api_keys(api_key_id) ON DELETE CASCADE;


--
-- Name: api_keys api_keys_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: audit_events audit_events_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_events
    ADD CONSTRAINT audit_events_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE SET NULL;


--
-- Name: auth_events auth_events_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_events
    ADD CONSTRAINT auth_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: auth_identities auth_identities_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_identities
    ADD CONSTRAINT auth_identities_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: auth_identity_versions auth_identity_versions_changed_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_identity_versions
    ADD CONSTRAINT auth_identity_versions_changed_by_user_id_fkey FOREIGN KEY (changed_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: automation_rules automation_rules_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.automation_rules
    ADD CONSTRAINT automation_rules_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: automation_rules automation_rules_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.automation_rules
    ADD CONSTRAINT automation_rules_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: automation_rules automation_rules_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.automation_rules
    ADD CONSTRAINT automation_rules_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: bot_sessions bot_sessions_action_dependency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_sessions
    ADD CONSTRAINT bot_sessions_action_dependency_id_fkey FOREIGN KEY (action_dependency_id) REFERENCES public.action_dependencies(action_dependency_id) ON DELETE CASCADE;


--
-- Name: bot_sessions bot_sessions_last_telegram_relay_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_sessions
    ADD CONSTRAINT bot_sessions_last_telegram_relay_event_id_fkey FOREIGN KEY (last_telegram_relay_event_id) REFERENCES public.telegram_relay_events(telegram_relay_event_id) ON DELETE SET NULL;


--
-- Name: bot_sessions bot_sessions_last_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_sessions
    ADD CONSTRAINT bot_sessions_last_workflow_run_id_fkey FOREIGN KEY (last_workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE SET NULL;


--
-- Name: bot_sessions bot_sessions_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_sessions
    ADD CONSTRAINT bot_sessions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: client_contacts client_contacts_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_contacts
    ADD CONSTRAINT client_contacts_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE CASCADE;


--
-- Name: client_contacts client_contacts_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_contacts
    ADD CONSTRAINT client_contacts_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: client_projects client_projects_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_projects
    ADD CONSTRAINT client_projects_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE CASCADE;


--
-- Name: client_projects client_projects_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_projects
    ADD CONSTRAINT client_projects_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: client_projects client_projects_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_projects
    ADD CONSTRAINT client_projects_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: client_responsibles client_responsibles_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_responsibles
    ADD CONSTRAINT client_responsibles_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE CASCADE;


--
-- Name: client_responsibles client_responsibles_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_responsibles
    ADD CONSTRAINT client_responsibles_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: client_responsibles client_responsibles_company_legal_entity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_responsibles
    ADD CONSTRAINT client_responsibles_company_legal_entity_id_fkey FOREIGN KEY (company_legal_entity_id) REFERENCES public.company_legal_entities(company_legal_entity_id) ON DELETE SET NULL;


--
-- Name: client_responsibles client_responsibles_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_responsibles
    ADD CONSTRAINT client_responsibles_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: client_responsibles client_responsibles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_responsibles
    ADD CONSTRAINT client_responsibles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: client_service_types client_service_types_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_service_types
    ADD CONSTRAINT client_service_types_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: client_service_types client_service_types_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_service_types
    ADD CONSTRAINT client_service_types_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: client_services client_services_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_services
    ADD CONSTRAINT client_services_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE CASCADE;


--
-- Name: client_services client_services_client_service_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_services
    ADD CONSTRAINT client_services_client_service_type_id_fkey FOREIGN KEY (client_service_type_id) REFERENCES public.client_service_types(client_service_type_id) ON DELETE RESTRICT;


--
-- Name: client_services client_services_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_services
    ADD CONSTRAINT client_services_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: client_services client_services_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_services
    ADD CONSTRAINT client_services_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: client_services client_services_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.client_services
    ADD CONSTRAINT client_services_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: clients clients_account_manager_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clients
    ADD CONSTRAINT clients_account_manager_user_id_fkey FOREIGN KEY (account_manager_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: clients clients_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clients
    ADD CONSTRAINT clients_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: clients clients_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clients
    ADD CONSTRAINT clients_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: clients clients_manager_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clients
    ADD CONSTRAINT clients_manager_user_id_fkey FOREIGN KEY (manager_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: companies companies_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.companies
    ADD CONSTRAINT companies_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: companies companies_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.companies
    ADD CONSTRAINT companies_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(user_id) ON DELETE RESTRICT;


--
-- Name: company_legal_entities company_legal_entities_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_legal_entities
    ADD CONSTRAINT company_legal_entities_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: company_legal_entities company_legal_entities_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_legal_entities
    ADD CONSTRAINT company_legal_entities_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: company_membership_versions company_membership_versions_changed_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_membership_versions
    ADD CONSTRAINT company_membership_versions_changed_by_user_id_fkey FOREIGN KEY (changed_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: company_memberships company_memberships_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_memberships
    ADD CONSTRAINT company_memberships_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: company_memberships company_memberships_invited_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_memberships
    ADD CONSTRAINT company_memberships_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: company_memberships company_memberships_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_memberships
    ADD CONSTRAINT company_memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: company_versions company_versions_changed_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.company_versions
    ADD CONSTRAINT company_versions_changed_by_user_id_fkey FOREIGN KEY (changed_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: desktop_auth_requests desktop_auth_requests_desktop_device_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_auth_requests
    ADD CONSTRAINT desktop_auth_requests_desktop_device_id_fkey FOREIGN KEY (desktop_device_id) REFERENCES public.desktop_devices(desktop_device_id) ON DELETE SET NULL;


--
-- Name: desktop_auth_requests desktop_auth_requests_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_auth_requests
    ADD CONSTRAINT desktop_auth_requests_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE SET NULL;


--
-- Name: desktop_auth_requests desktop_auth_requests_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_auth_requests
    ADD CONSTRAINT desktop_auth_requests_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: desktop_devices desktop_devices_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_devices
    ADD CONSTRAINT desktop_devices_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: desktop_devices desktop_devices_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_devices
    ADD CONSTRAINT desktop_devices_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: desktop_messages desktop_messages_desktop_device_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_messages
    ADD CONSTRAINT desktop_messages_desktop_device_id_fkey FOREIGN KEY (desktop_device_id) REFERENCES public.desktop_devices(desktop_device_id) ON DELETE SET NULL;


--
-- Name: desktop_messages desktop_messages_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_messages
    ADD CONSTRAINT desktop_messages_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: desktop_messages desktop_messages_target_webhook_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_messages
    ADD CONSTRAINT desktop_messages_target_webhook_event_id_fkey FOREIGN KEY (target_webhook_event_id) REFERENCES public.webhook_events(webhook_event_id) ON DELETE SET NULL;


--
-- Name: desktop_messages desktop_messages_target_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_messages
    ADD CONSTRAINT desktop_messages_target_workflow_id_fkey FOREIGN KEY (target_workflow_id) REFERENCES public.workflows(workflow_id) ON DELETE SET NULL;


--
-- Name: desktop_messages desktop_messages_target_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_messages
    ADD CONSTRAINT desktop_messages_target_workflow_run_id_fkey FOREIGN KEY (target_workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE SET NULL;


--
-- Name: desktop_tokens desktop_tokens_desktop_device_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.desktop_tokens
    ADD CONSTRAINT desktop_tokens_desktop_device_id_fkey FOREIGN KEY (desktop_device_id) REFERENCES public.desktop_devices(desktop_device_id) ON DELETE CASCADE;


--
-- Name: device_credential_versions device_credential_versions_changed_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_credential_versions
    ADD CONSTRAINT device_credential_versions_changed_by_user_id_fkey FOREIGN KEY (changed_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: device_credentials device_credentials_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_credentials
    ADD CONSTRAINT device_credentials_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.sessions(session_id) ON DELETE SET NULL;


--
-- Name: device_credentials device_credentials_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_credentials
    ADD CONSTRAINT device_credentials_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: effect_ledger effect_ledger_action_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.effect_ledger
    ADD CONSTRAINT effect_ledger_action_run_id_fkey FOREIGN KEY (action_run_id) REFERENCES public.action_runs(action_run_id) ON DELETE SET NULL;


--
-- Name: effect_ledger effect_ledger_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.effect_ledger
    ADD CONSTRAINT effect_ledger_workflow_run_id_fkey FOREIGN KEY (workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE CASCADE;


--
-- Name: email_templates email_templates_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_templates
    ADD CONSTRAINT email_templates_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: email_templates email_templates_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_templates
    ADD CONSTRAINT email_templates_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: employee_profiles employee_profiles_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_profiles
    ADD CONSTRAINT employee_profiles_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: employee_profiles employee_profiles_manager_employee_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_profiles
    ADD CONSTRAINT employee_profiles_manager_employee_profile_id_fkey FOREIGN KEY (manager_employee_profile_id) REFERENCES public.employee_profiles(employee_profile_id) ON DELETE SET NULL;


--
-- Name: employee_profiles employee_profiles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_profiles
    ADD CONSTRAINT employee_profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: integration_doc_chunks integration_doc_chunks_integration_doc_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.integration_doc_chunks
    ADD CONSTRAINT integration_doc_chunks_integration_doc_id_fkey FOREIGN KEY (integration_doc_id) REFERENCES public.integration_docs(integration_doc_id) ON DELETE CASCADE;


--
-- Name: message_templates message_templates_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message_templates
    ADD CONSTRAINT message_templates_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: message_templates message_templates_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message_templates
    ADD CONSTRAINT message_templates_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: module_channel_bindings module_channel_bindings_action_dependency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_channel_bindings
    ADD CONSTRAINT module_channel_bindings_action_dependency_id_fkey FOREIGN KEY (action_dependency_id) REFERENCES public.action_dependencies(action_dependency_id) ON DELETE SET NULL;


--
-- Name: module_channel_bindings module_channel_bindings_action_type_key_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_channel_bindings
    ADD CONSTRAINT module_channel_bindings_action_type_key_fkey FOREIGN KEY (action_type_key) REFERENCES public.action_types(action_key) ON DELETE RESTRICT;


--
-- Name: module_channel_bindings module_channel_bindings_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_channel_bindings
    ADD CONSTRAINT module_channel_bindings_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: module_channel_bindings module_channel_bindings_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_channel_bindings
    ADD CONSTRAINT module_channel_bindings_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: module_channel_bindings module_channel_bindings_email_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_channel_bindings
    ADD CONSTRAINT module_channel_bindings_email_template_id_fkey FOREIGN KEY (email_template_id) REFERENCES public.email_templates(email_template_id) ON DELETE SET NULL;


--
-- Name: module_channel_bindings module_channel_bindings_message_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_channel_bindings
    ADD CONSTRAINT module_channel_bindings_message_template_id_fkey FOREIGN KEY (message_template_id) REFERENCES public.message_templates(message_template_id) ON DELETE SET NULL;


--
-- Name: module_channel_bindings module_channel_bindings_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_channel_bindings
    ADD CONSTRAINT module_channel_bindings_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: module_workflow_triggers module_workflow_triggers_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_workflow_triggers
    ADD CONSTRAINT module_workflow_triggers_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: module_workflow_triggers module_workflow_triggers_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_workflow_triggers
    ADD CONSTRAINT module_workflow_triggers_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: module_workflow_triggers module_workflow_triggers_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_workflow_triggers
    ADD CONSTRAINT module_workflow_triggers_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: module_workflow_triggers module_workflow_triggers_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_workflow_triggers
    ADD CONSTRAINT module_workflow_triggers_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES public.workflows(workflow_id) ON DELETE CASCADE;


--
-- Name: module_workflow_triggers module_workflow_triggers_workflow_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_workflow_triggers
    ADD CONSTRAINT module_workflow_triggers_workflow_version_id_fkey FOREIGN KEY (workflow_version_id) REFERENCES public.workflow_versions(workflow_version_id) ON DELETE SET NULL;


--
-- Name: notifications notifications_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: notifications notifications_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(task_id) ON DELETE CASCADE;


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: payment_action_documents payment_action_documents_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_action_documents
    ADD CONSTRAINT payment_action_documents_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE CASCADE;


--
-- Name: payment_action_documents payment_action_documents_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_action_documents
    ADD CONSTRAINT payment_action_documents_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: payment_action_documents payment_action_documents_payment_action_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_action_documents
    ADD CONSTRAINT payment_action_documents_payment_action_id_fkey FOREIGN KEY (payment_action_id) REFERENCES public.payment_actions(payment_action_id) ON DELETE CASCADE;


--
-- Name: payment_action_documents payment_action_documents_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_action_documents
    ADD CONSTRAINT payment_action_documents_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: payment_action_documents payment_action_documents_uploaded_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_action_documents
    ADD CONSTRAINT payment_action_documents_uploaded_by_user_id_fkey FOREIGN KEY (uploaded_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: payment_action_items payment_action_items_client_service_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_action_items
    ADD CONSTRAINT payment_action_items_client_service_id_fkey FOREIGN KEY (client_service_id) REFERENCES public.client_services(client_service_id) ON DELETE RESTRICT;


--
-- Name: payment_action_items payment_action_items_client_service_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_action_items
    ADD CONSTRAINT payment_action_items_client_service_type_id_fkey FOREIGN KEY (client_service_type_id) REFERENCES public.client_service_types(client_service_type_id) ON DELETE SET NULL;


--
-- Name: payment_action_items payment_action_items_payment_action_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_action_items
    ADD CONSTRAINT payment_action_items_payment_action_id_fkey FOREIGN KEY (payment_action_id) REFERENCES public.payment_actions(payment_action_id) ON DELETE CASCADE;


--
-- Name: payment_actions payment_actions_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_actions
    ADD CONSTRAINT payment_actions_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE CASCADE;


--
-- Name: payment_actions payment_actions_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_actions
    ADD CONSTRAINT payment_actions_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: payment_actions payment_actions_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_actions
    ADD CONSTRAINT payment_actions_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: payment_actions payment_actions_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_actions
    ADD CONSTRAINT payment_actions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: payment_actions payment_actions_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_actions
    ADD CONSTRAINT payment_actions_workflow_run_id_fkey FOREIGN KEY (workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE SET NULL;


--
-- Name: payment_events payment_events_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_events
    ADD CONSTRAINT payment_events_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE CASCADE;


--
-- Name: payment_events payment_events_client_service_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_events
    ADD CONSTRAINT payment_events_client_service_id_fkey FOREIGN KEY (client_service_id) REFERENCES public.client_services(client_service_id) ON DELETE SET NULL;


--
-- Name: payment_events payment_events_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_events
    ADD CONSTRAINT payment_events_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: payment_events payment_events_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_events
    ADD CONSTRAINT payment_events_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: payment_events payment_events_payment_action_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_events
    ADD CONSTRAINT payment_events_payment_action_id_fkey FOREIGN KEY (payment_action_id) REFERENCES public.payment_actions(payment_action_id) ON DELETE SET NULL;


--
-- Name: payment_events payment_events_payment_fact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_events
    ADD CONSTRAINT payment_events_payment_fact_id_fkey FOREIGN KEY (payment_fact_id) REFERENCES public.payment_facts(payment_fact_id) ON DELETE SET NULL;


--
-- Name: payment_events payment_events_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_events
    ADD CONSTRAINT payment_events_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: payment_events payment_events_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_events
    ADD CONSTRAINT payment_events_workflow_run_id_fkey FOREIGN KEY (workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE SET NULL;


--
-- Name: payment_facts payment_facts_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_facts
    ADD CONSTRAINT payment_facts_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE CASCADE;


--
-- Name: payment_facts payment_facts_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_facts
    ADD CONSTRAINT payment_facts_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: payment_facts payment_facts_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_facts
    ADD CONSTRAINT payment_facts_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: payment_facts payment_facts_payment_action_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_facts
    ADD CONSTRAINT payment_facts_payment_action_id_fkey FOREIGN KEY (payment_action_id) REFERENCES public.payment_actions(payment_action_id) ON DELETE CASCADE;


--
-- Name: payment_facts payment_facts_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_facts
    ADD CONSTRAINT payment_facts_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: project_ai_employees project_ai_employees_action_dependency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_ai_employees
    ADD CONSTRAINT project_ai_employees_action_dependency_id_fkey FOREIGN KEY (action_dependency_id) REFERENCES public.action_dependencies(action_dependency_id) ON DELETE SET NULL;


--
-- Name: project_ai_employees project_ai_employees_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_ai_employees
    ADD CONSTRAINT project_ai_employees_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: project_ai_employees project_ai_employees_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_ai_employees
    ADD CONSTRAINT project_ai_employees_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: project_ai_employees project_ai_employees_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_ai_employees
    ADD CONSTRAINT project_ai_employees_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: project_member_versions project_member_versions_changed_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_member_versions
    ADD CONSTRAINT project_member_versions_changed_by_user_id_fkey FOREIGN KEY (changed_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: project_members project_members_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE SET NULL;


--
-- Name: project_members project_members_invited_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_invited_by_user_id_fkey FOREIGN KEY (invited_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: project_members project_members_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: project_members project_members_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: project_versions project_versions_changed_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_versions
    ADD CONSTRAINT project_versions_changed_by_user_id_fkey FOREIGN KEY (changed_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: projects projects_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: projects projects_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: projects projects_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: projects projects_project_manager_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_project_manager_user_id_fkey FOREIGN KEY (project_manager_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: projects projects_task_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_task_workflow_id_fkey FOREIGN KEY (task_workflow_id) REFERENCES public.task_workflows(task_workflow_id) ON DELETE SET NULL;


--
-- Name: provider_operation_effects provider_operation_effects_action_dependency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.provider_operation_effects
    ADD CONSTRAINT provider_operation_effects_action_dependency_id_fkey FOREIGN KEY (action_dependency_id) REFERENCES public.action_dependencies(action_dependency_id) ON DELETE RESTRICT;


--
-- Name: run_events run_events_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.run_events
    ADD CONSTRAINT run_events_workflow_run_id_fkey FOREIGN KEY (workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE CASCADE;


--
-- Name: secrets secrets_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.secrets
    ADD CONSTRAINT secrets_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: secrets secrets_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.secrets
    ADD CONSTRAINT secrets_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: secrets secrets_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.secrets
    ADD CONSTRAINT secrets_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: site_call_sessions site_call_sessions_chat_widget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_call_sessions
    ADD CONSTRAINT site_call_sessions_chat_widget_id_fkey FOREIGN KEY (chat_widget_id) REFERENCES public.site_chat_widgets(chat_widget_id) ON DELETE CASCADE;


--
-- Name: site_call_sessions site_call_sessions_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_call_sessions
    ADD CONSTRAINT site_call_sessions_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: site_call_sessions site_call_sessions_conversation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_call_sessions
    ADD CONSTRAINT site_call_sessions_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES public.site_chat_conversations(conversation_id) ON DELETE SET NULL;


--
-- Name: site_call_sessions site_call_sessions_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_call_sessions
    ADD CONSTRAINT site_call_sessions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: site_call_sessions site_call_sessions_requested_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_call_sessions
    ADD CONSTRAINT site_call_sessions_requested_by_user_id_fkey FOREIGN KEY (requested_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: site_call_sessions site_call_sessions_visit_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_call_sessions
    ADD CONSTRAINT site_call_sessions_visit_session_id_fkey FOREIGN KEY (visit_session_id) REFERENCES public.site_visit_sessions(visit_session_id) ON DELETE SET NULL;


--
-- Name: site_call_sessions site_call_sessions_visitor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_call_sessions
    ADD CONSTRAINT site_call_sessions_visitor_id_fkey FOREIGN KEY (visitor_id) REFERENCES public.site_visitors(visitor_id) ON DELETE CASCADE;


--
-- Name: site_chat_attachments site_chat_attachments_call_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_attachments
    ADD CONSTRAINT site_chat_attachments_call_session_id_fkey FOREIGN KEY (call_session_id) REFERENCES public.site_call_sessions(call_session_id) ON DELETE SET NULL;


--
-- Name: site_chat_attachments site_chat_attachments_chat_widget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_attachments
    ADD CONSTRAINT site_chat_attachments_chat_widget_id_fkey FOREIGN KEY (chat_widget_id) REFERENCES public.site_chat_widgets(chat_widget_id) ON DELETE SET NULL;


--
-- Name: site_chat_attachments site_chat_attachments_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_attachments
    ADD CONSTRAINT site_chat_attachments_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: site_chat_attachments site_chat_attachments_conversation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_attachments
    ADD CONSTRAINT site_chat_attachments_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES public.site_chat_conversations(conversation_id) ON DELETE CASCADE;


--
-- Name: site_chat_attachments site_chat_attachments_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_attachments
    ADD CONSTRAINT site_chat_attachments_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.site_chat_messages(message_id) ON DELETE SET NULL;


--
-- Name: site_chat_attachments site_chat_attachments_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_attachments
    ADD CONSTRAINT site_chat_attachments_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: site_chat_attachments site_chat_attachments_uploaded_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_attachments
    ADD CONSTRAINT site_chat_attachments_uploaded_by_user_id_fkey FOREIGN KEY (uploaded_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: site_chat_conversations site_chat_conversations_chat_widget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_conversations
    ADD CONSTRAINT site_chat_conversations_chat_widget_id_fkey FOREIGN KEY (chat_widget_id) REFERENCES public.site_chat_widgets(chat_widget_id) ON DELETE CASCADE;


--
-- Name: site_chat_conversations site_chat_conversations_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_conversations
    ADD CONSTRAINT site_chat_conversations_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: site_chat_conversations site_chat_conversations_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_conversations
    ADD CONSTRAINT site_chat_conversations_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: site_chat_conversations site_chat_conversations_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_conversations
    ADD CONSTRAINT site_chat_conversations_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: site_chat_conversations site_chat_conversations_visit_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_conversations
    ADD CONSTRAINT site_chat_conversations_visit_session_id_fkey FOREIGN KEY (visit_session_id) REFERENCES public.site_visit_sessions(visit_session_id) ON DELETE SET NULL;


--
-- Name: site_chat_conversations site_chat_conversations_visitor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_conversations
    ADD CONSTRAINT site_chat_conversations_visitor_id_fkey FOREIGN KEY (visitor_id) REFERENCES public.site_visitors(visitor_id) ON DELETE CASCADE;


--
-- Name: site_chat_messages site_chat_messages_author_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_messages
    ADD CONSTRAINT site_chat_messages_author_user_id_fkey FOREIGN KEY (author_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: site_chat_messages site_chat_messages_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_messages
    ADD CONSTRAINT site_chat_messages_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: site_chat_messages site_chat_messages_conversation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_messages
    ADD CONSTRAINT site_chat_messages_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES public.site_chat_conversations(conversation_id) ON DELETE CASCADE;


--
-- Name: site_chat_messages site_chat_messages_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_messages
    ADD CONSTRAINT site_chat_messages_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: site_chat_widgets site_chat_widgets_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_widgets
    ADD CONSTRAINT site_chat_widgets_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: site_chat_widgets site_chat_widgets_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_widgets
    ADD CONSTRAINT site_chat_widgets_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: site_chat_widgets site_chat_widgets_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_chat_widgets
    ADD CONSTRAINT site_chat_widgets_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: site_cobrowse_frames site_cobrowse_frames_chat_widget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_cobrowse_frames
    ADD CONSTRAINT site_cobrowse_frames_chat_widget_id_fkey FOREIGN KEY (chat_widget_id) REFERENCES public.site_chat_widgets(chat_widget_id) ON DELETE CASCADE;


--
-- Name: site_cobrowse_frames site_cobrowse_frames_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_cobrowse_frames
    ADD CONSTRAINT site_cobrowse_frames_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: site_cobrowse_frames site_cobrowse_frames_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_cobrowse_frames
    ADD CONSTRAINT site_cobrowse_frames_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: site_cobrowse_frames site_cobrowse_frames_visit_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_cobrowse_frames
    ADD CONSTRAINT site_cobrowse_frames_visit_session_id_fkey FOREIGN KEY (visit_session_id) REFERENCES public.site_visit_sessions(visit_session_id) ON DELETE CASCADE;


--
-- Name: site_cobrowse_frames site_cobrowse_frames_visitor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_cobrowse_frames
    ADD CONSTRAINT site_cobrowse_frames_visitor_id_fkey FOREIGN KEY (visitor_id) REFERENCES public.site_visitors(visitor_id) ON DELETE CASCADE;


--
-- Name: site_visit_connections site_visit_connections_chat_widget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visit_connections
    ADD CONSTRAINT site_visit_connections_chat_widget_id_fkey FOREIGN KEY (chat_widget_id) REFERENCES public.site_chat_widgets(chat_widget_id) ON DELETE CASCADE;


--
-- Name: site_visit_connections site_visit_connections_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visit_connections
    ADD CONSTRAINT site_visit_connections_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: site_visit_connections site_visit_connections_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visit_connections
    ADD CONSTRAINT site_visit_connections_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: site_visit_connections site_visit_connections_visit_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visit_connections
    ADD CONSTRAINT site_visit_connections_visit_session_id_fkey FOREIGN KEY (visit_session_id) REFERENCES public.site_visit_sessions(visit_session_id) ON DELETE SET NULL;


--
-- Name: site_visit_connections site_visit_connections_visitor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visit_connections
    ADD CONSTRAINT site_visit_connections_visitor_id_fkey FOREIGN KEY (visitor_id) REFERENCES public.site_visitors(visitor_id) ON DELETE CASCADE;


--
-- Name: site_visit_sessions site_visit_sessions_chat_widget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visit_sessions
    ADD CONSTRAINT site_visit_sessions_chat_widget_id_fkey FOREIGN KEY (chat_widget_id) REFERENCES public.site_chat_widgets(chat_widget_id) ON DELETE CASCADE;


--
-- Name: site_visit_sessions site_visit_sessions_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visit_sessions
    ADD CONSTRAINT site_visit_sessions_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: site_visit_sessions site_visit_sessions_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visit_sessions
    ADD CONSTRAINT site_visit_sessions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: site_visit_sessions site_visit_sessions_visitor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visit_sessions
    ADD CONSTRAINT site_visit_sessions_visitor_id_fkey FOREIGN KEY (visitor_id) REFERENCES public.site_visitors(visitor_id) ON DELETE CASCADE;


--
-- Name: site_visitor_events site_visitor_events_chat_widget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visitor_events
    ADD CONSTRAINT site_visitor_events_chat_widget_id_fkey FOREIGN KEY (chat_widget_id) REFERENCES public.site_chat_widgets(chat_widget_id) ON DELETE CASCADE;


--
-- Name: site_visitor_events site_visitor_events_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visitor_events
    ADD CONSTRAINT site_visitor_events_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: site_visitor_events site_visitor_events_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visitor_events
    ADD CONSTRAINT site_visitor_events_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: site_visitor_events site_visitor_events_visit_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visitor_events
    ADD CONSTRAINT site_visitor_events_visit_session_id_fkey FOREIGN KEY (visit_session_id) REFERENCES public.site_visit_sessions(visit_session_id) ON DELETE CASCADE;


--
-- Name: site_visitor_events site_visitor_events_visitor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visitor_events
    ADD CONSTRAINT site_visitor_events_visitor_id_fkey FOREIGN KEY (visitor_id) REFERENCES public.site_visitors(visitor_id) ON DELETE CASCADE;


--
-- Name: site_visitors site_visitors_chat_widget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visitors
    ADD CONSTRAINT site_visitors_chat_widget_id_fkey FOREIGN KEY (chat_widget_id) REFERENCES public.site_chat_widgets(chat_widget_id) ON DELETE CASCADE;


--
-- Name: site_visitors site_visitors_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visitors
    ADD CONSTRAINT site_visitors_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE SET NULL;


--
-- Name: site_visitors site_visitors_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visitors
    ADD CONSTRAINT site_visitors_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: site_visitors site_visitors_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.site_visitors
    ADD CONSTRAINT site_visitors_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: task_activity_events task_activity_events_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_activity_events
    ADD CONSTRAINT task_activity_events_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: task_activity_events task_activity_events_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_activity_events
    ADD CONSTRAINT task_activity_events_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: task_activity_events task_activity_events_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_activity_events
    ADD CONSTRAINT task_activity_events_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(task_id) ON DELETE CASCADE;


--
-- Name: task_attachments task_attachments_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE SET NULL;


--
-- Name: task_attachments task_attachments_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: task_attachments task_attachments_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: task_attachments task_attachments_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(task_id) ON DELETE CASCADE;


--
-- Name: task_attachments task_attachments_uploaded_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_uploaded_by_user_id_fkey FOREIGN KEY (uploaded_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: task_boards task_boards_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_boards
    ADD CONSTRAINT task_boards_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE SET NULL;


--
-- Name: task_boards task_boards_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_boards
    ADD CONSTRAINT task_boards_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: task_boards task_boards_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_boards
    ADD CONSTRAINT task_boards_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: task_boards task_boards_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_boards
    ADD CONSTRAINT task_boards_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: task_boards task_boards_task_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_boards
    ADD CONSTRAINT task_boards_task_workflow_id_fkey FOREIGN KEY (task_workflow_id) REFERENCES public.task_workflows(task_workflow_id) ON DELETE CASCADE;


--
-- Name: task_checklist_items task_checklist_items_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_checklist_items
    ADD CONSTRAINT task_checklist_items_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(task_id) ON DELETE CASCADE;


--
-- Name: task_comments task_comments_author_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_author_user_id_fkey FOREIGN KEY (author_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: task_comments task_comments_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: task_comments task_comments_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(task_id) ON DELETE CASCADE;


--
-- Name: task_custom_field_values task_custom_field_values_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_custom_field_values
    ADD CONSTRAINT task_custom_field_values_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE CASCADE;


--
-- Name: task_custom_field_values task_custom_field_values_employee_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_custom_field_values
    ADD CONSTRAINT task_custom_field_values_employee_profile_id_fkey FOREIGN KEY (employee_profile_id) REFERENCES public.employee_profiles(employee_profile_id) ON DELETE CASCADE;


--
-- Name: task_custom_field_values task_custom_field_values_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_custom_field_values
    ADD CONSTRAINT task_custom_field_values_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: task_custom_field_values task_custom_field_values_task_custom_field_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_custom_field_values
    ADD CONSTRAINT task_custom_field_values_task_custom_field_id_fkey FOREIGN KEY (task_custom_field_id) REFERENCES public.task_custom_fields(task_custom_field_id) ON DELETE CASCADE;


--
-- Name: task_custom_field_values task_custom_field_values_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_custom_field_values
    ADD CONSTRAINT task_custom_field_values_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(task_id) ON DELETE CASCADE;


--
-- Name: task_custom_fields task_custom_fields_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_custom_fields
    ADD CONSTRAINT task_custom_fields_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: task_custom_fields task_custom_fields_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_custom_fields
    ADD CONSTRAINT task_custom_fields_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: task_custom_fields task_custom_fields_task_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_custom_fields
    ADD CONSTRAINT task_custom_fields_task_workflow_id_fkey FOREIGN KEY (task_workflow_id) REFERENCES public.task_workflows(task_workflow_id) ON DELETE CASCADE;


--
-- Name: task_process_memberships task_process_memberships_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_process_memberships
    ADD CONSTRAINT task_process_memberships_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE SET NULL;


--
-- Name: task_process_memberships task_process_memberships_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_process_memberships
    ADD CONSTRAINT task_process_memberships_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: task_process_memberships task_process_memberships_invited_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_process_memberships
    ADD CONSTRAINT task_process_memberships_invited_by_user_id_fkey FOREIGN KEY (invited_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: task_process_memberships task_process_memberships_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_process_memberships
    ADD CONSTRAINT task_process_memberships_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: task_process_memberships task_process_memberships_task_process_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_process_memberships
    ADD CONSTRAINT task_process_memberships_task_process_role_id_fkey FOREIGN KEY (task_process_role_id) REFERENCES public.task_process_roles(task_process_role_id) ON DELETE RESTRICT;


--
-- Name: task_process_memberships task_process_memberships_task_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_process_memberships
    ADD CONSTRAINT task_process_memberships_task_workflow_id_fkey FOREIGN KEY (task_workflow_id) REFERENCES public.task_workflows(task_workflow_id) ON DELETE CASCADE;


--
-- Name: task_process_memberships task_process_memberships_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_process_memberships
    ADD CONSTRAINT task_process_memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: task_process_roles task_process_roles_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_process_roles
    ADD CONSTRAINT task_process_roles_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: task_process_roles task_process_roles_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_process_roles
    ADD CONSTRAINT task_process_roles_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: task_process_roles task_process_roles_task_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_process_roles
    ADD CONSTRAINT task_process_roles_task_workflow_id_fkey FOREIGN KEY (task_workflow_id) REFERENCES public.task_workflows(task_workflow_id) ON DELETE CASCADE;


--
-- Name: task_status_history task_status_history_changed_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_status_history
    ADD CONSTRAINT task_status_history_changed_by_user_id_fkey FOREIGN KEY (changed_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: task_status_history task_status_history_from_status_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_status_history
    ADD CONSTRAINT task_status_history_from_status_id_fkey FOREIGN KEY (from_status_id) REFERENCES public.task_workflow_statuses(task_workflow_status_id) ON DELETE SET NULL;


--
-- Name: task_status_history task_status_history_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_status_history
    ADD CONSTRAINT task_status_history_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(task_id) ON DELETE CASCADE;


--
-- Name: task_status_history task_status_history_to_status_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_status_history
    ADD CONSTRAINT task_status_history_to_status_id_fkey FOREIGN KEY (to_status_id) REFERENCES public.task_workflow_statuses(task_workflow_status_id) ON DELETE RESTRICT;


--
-- Name: task_views task_views_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_views
    ADD CONSTRAINT task_views_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: task_views task_views_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_views
    ADD CONSTRAINT task_views_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: task_views task_views_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_views
    ADD CONSTRAINT task_views_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: task_watchers task_watchers_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_watchers
    ADD CONSTRAINT task_watchers_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(task_id) ON DELETE CASCADE;


--
-- Name: task_watchers task_watchers_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_watchers
    ADD CONSTRAINT task_watchers_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: task_workflow_statuses task_workflow_statuses_task_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_workflow_statuses
    ADD CONSTRAINT task_workflow_statuses_task_workflow_id_fkey FOREIGN KEY (task_workflow_id) REFERENCES public.task_workflows(task_workflow_id) ON DELETE CASCADE;


--
-- Name: task_workflow_transitions task_workflow_transitions_from_status_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_workflow_transitions
    ADD CONSTRAINT task_workflow_transitions_from_status_id_fkey FOREIGN KEY (from_status_id) REFERENCES public.task_workflow_statuses(task_workflow_status_id) ON DELETE CASCADE;


--
-- Name: task_workflow_transitions task_workflow_transitions_task_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_workflow_transitions
    ADD CONSTRAINT task_workflow_transitions_task_workflow_id_fkey FOREIGN KEY (task_workflow_id) REFERENCES public.task_workflows(task_workflow_id) ON DELETE CASCADE;


--
-- Name: task_workflow_transitions task_workflow_transitions_to_status_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_workflow_transitions
    ADD CONSTRAINT task_workflow_transitions_to_status_id_fkey FOREIGN KEY (to_status_id) REFERENCES public.task_workflow_statuses(task_workflow_status_id) ON DELETE CASCADE;


--
-- Name: task_workflows task_workflows_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_workflows
    ADD CONSTRAINT task_workflows_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE SET NULL;


--
-- Name: task_workflows task_workflows_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_workflows
    ADD CONSTRAINT task_workflows_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: task_workflows task_workflows_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_workflows
    ADD CONSTRAINT task_workflows_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: task_workflows task_workflows_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_workflows
    ADD CONSTRAINT task_workflows_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: tasks tasks_assignee_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_assignee_user_id_fkey FOREIGN KEY (assignee_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: tasks tasks_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE SET NULL;


--
-- Name: tasks tasks_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: tasks tasks_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: tasks tasks_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: tasks tasks_parent_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_parent_task_id_fkey FOREIGN KEY (parent_task_id) REFERENCES public.tasks(task_id) ON DELETE SET NULL;


--
-- Name: tasks tasks_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE SET NULL;


--
-- Name: tasks tasks_reporter_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_reporter_user_id_fkey FOREIGN KEY (reporter_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: tasks tasks_reviewer_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_reviewer_user_id_fkey FOREIGN KEY (reviewer_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: tasks tasks_task_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_task_workflow_id_fkey FOREIGN KEY (task_workflow_id) REFERENCES public.task_workflows(task_workflow_id) ON DELETE RESTRICT;


--
-- Name: tasks tasks_task_workflow_status_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_task_workflow_status_id_fkey FOREIGN KEY (task_workflow_status_id) REFERENCES public.task_workflow_statuses(task_workflow_status_id) ON DELETE RESTRICT;


--
-- Name: team_members team_members_employee_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_employee_profile_id_fkey FOREIGN KEY (employee_profile_id) REFERENCES public.employee_profiles(employee_profile_id) ON DELETE CASCADE;


--
-- Name: team_members team_members_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE;


--
-- Name: teams teams_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: teams teams_lead_employee_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_lead_employee_profile_id_fkey FOREIGN KEY (lead_employee_profile_id) REFERENCES public.employee_profiles(employee_profile_id) ON DELETE SET NULL;


--
-- Name: telegram_relay_configs telegram_relay_configs_action_dependency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.telegram_relay_configs
    ADD CONSTRAINT telegram_relay_configs_action_dependency_id_fkey FOREIGN KEY (action_dependency_id) REFERENCES public.action_dependencies(action_dependency_id) ON DELETE CASCADE;


--
-- Name: telegram_relay_events telegram_relay_events_action_dependency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.telegram_relay_events
    ADD CONSTRAINT telegram_relay_events_action_dependency_id_fkey FOREIGN KEY (action_dependency_id) REFERENCES public.action_dependencies(action_dependency_id) ON DELETE CASCADE;


--
-- Name: telegram_relay_events telegram_relay_events_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.telegram_relay_events
    ADD CONSTRAINT telegram_relay_events_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: telegram_relay_forwards telegram_relay_forwards_telegram_relay_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.telegram_relay_forwards
    ADD CONSTRAINT telegram_relay_forwards_telegram_relay_event_id_fkey FOREIGN KEY (telegram_relay_event_id) REFERENCES public.telegram_relay_events(telegram_relay_event_id) ON DELETE CASCADE;


--
-- Name: time_entries time_entries_approved_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_approved_by_user_id_fkey FOREIGN KEY (approved_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: time_entries time_entries_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(client_id) ON DELETE SET NULL;


--
-- Name: time_entries time_entries_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: time_entries time_entries_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE SET NULL;


--
-- Name: time_entries time_entries_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(task_id) ON DELETE SET NULL;


--
-- Name: time_entries time_entries_task_workflow_status_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_task_workflow_status_id_fkey FOREIGN KEY (task_workflow_status_id) REFERENCES public.task_workflow_statuses(task_workflow_status_id) ON DELETE SET NULL;


--
-- Name: time_entries time_entries_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: timer_sessions timer_sessions_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timer_sessions
    ADD CONSTRAINT timer_sessions_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.companies(company_id) ON DELETE CASCADE;


--
-- Name: timer_sessions timer_sessions_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timer_sessions
    ADD CONSTRAINT timer_sessions_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(task_id) ON DELETE CASCADE;


--
-- Name: timer_sessions timer_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timer_sessions
    ADD CONSTRAINT timer_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: user_versions user_versions_changed_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_versions
    ADD CONSTRAINT user_versions_changed_by_user_id_fkey FOREIGN KEY (changed_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: user_workspace_contexts user_workspace_contexts_active_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_workspace_contexts
    ADD CONSTRAINT user_workspace_contexts_active_company_id_fkey FOREIGN KEY (active_company_id) REFERENCES public.companies(company_id) ON DELETE SET NULL;


--
-- Name: user_workspace_contexts user_workspace_contexts_active_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_workspace_contexts
    ADD CONSTRAINT user_workspace_contexts_active_project_id_fkey FOREIGN KEY (active_project_id) REFERENCES public.projects(project_id) ON DELETE SET NULL;


--
-- Name: user_workspace_contexts user_workspace_contexts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_workspace_contexts
    ADD CONSTRAINT user_workspace_contexts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: users users_current_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_current_company_id_fkey FOREIGN KEY (current_company_id) REFERENCES public.companies(company_id) ON DELETE SET NULL;


--
-- Name: webhook_endpoints webhook_endpoints_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_endpoints
    ADD CONSTRAINT webhook_endpoints_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: webhook_endpoints webhook_endpoints_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_endpoints
    ADD CONSTRAINT webhook_endpoints_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: webhook_endpoints webhook_endpoints_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_endpoints
    ADD CONSTRAINT webhook_endpoints_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES public.workflows(workflow_id) ON DELETE CASCADE;


--
-- Name: webhook_endpoints webhook_endpoints_workflow_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_endpoints
    ADD CONSTRAINT webhook_endpoints_workflow_version_id_fkey FOREIGN KEY (workflow_version_id) REFERENCES public.workflow_versions(workflow_version_id) ON DELETE RESTRICT;


--
-- Name: webhook_events webhook_events_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: webhook_events webhook_events_webhook_endpoint_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_webhook_endpoint_id_fkey FOREIGN KEY (webhook_endpoint_id) REFERENCES public.webhook_endpoints(webhook_endpoint_id) ON DELETE CASCADE;


--
-- Name: webhook_events webhook_events_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES public.workflows(workflow_id) ON DELETE CASCADE;


--
-- Name: webhook_events webhook_events_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_workflow_run_id_fkey FOREIGN KEY (workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE SET NULL;


--
-- Name: webhook_events webhook_events_workflow_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_workflow_version_id_fkey FOREIGN KEY (workflow_version_id) REFERENCES public.workflow_versions(workflow_version_id) ON DELETE RESTRICT;


--
-- Name: workflow_action_dead_letters workflow_action_dead_letters_action_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_action_dead_letters
    ADD CONSTRAINT workflow_action_dead_letters_action_run_id_fkey FOREIGN KEY (action_run_id) REFERENCES public.action_runs(action_run_id) ON DELETE CASCADE;


--
-- Name: workflow_action_dead_letters workflow_action_dead_letters_replay_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_action_dead_letters
    ADD CONSTRAINT workflow_action_dead_letters_replay_workflow_run_id_fkey FOREIGN KEY (replay_workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE SET NULL;


--
-- Name: workflow_action_dead_letters workflow_action_dead_letters_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_action_dead_letters
    ADD CONSTRAINT workflow_action_dead_letters_workflow_run_id_fkey FOREIGN KEY (workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE CASCADE;


--
-- Name: workflow_client_payment_due_state workflow_client_payment_due_state_last_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_client_payment_due_state
    ADD CONSTRAINT workflow_client_payment_due_state_last_workflow_run_id_fkey FOREIGN KEY (last_workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE SET NULL;


--
-- Name: workflow_client_payment_due_state workflow_client_payment_due_state_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_client_payment_due_state
    ADD CONSTRAINT workflow_client_payment_due_state_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: workflow_client_payment_due_state workflow_client_payment_due_state_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_client_payment_due_state
    ADD CONSTRAINT workflow_client_payment_due_state_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES public.workflows(workflow_id) ON DELETE CASCADE;


--
-- Name: workflow_client_payment_due_state workflow_client_payment_due_state_workflow_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_client_payment_due_state
    ADD CONSTRAINT workflow_client_payment_due_state_workflow_version_id_fkey FOREIGN KEY (workflow_version_id) REFERENCES public.workflow_versions(workflow_version_id) ON DELETE RESTRICT;


--
-- Name: workflow_runs workflow_runs_cancel_requested_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_runs
    ADD CONSTRAINT workflow_runs_cancel_requested_by_user_id_fkey FOREIGN KEY (cancel_requested_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: workflow_runs workflow_runs_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_runs
    ADD CONSTRAINT workflow_runs_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: workflow_runs workflow_runs_replayed_from_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_runs
    ADD CONSTRAINT workflow_runs_replayed_from_workflow_run_id_fkey FOREIGN KEY (replayed_from_workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE SET NULL;


--
-- Name: workflow_runs workflow_runs_source_webhook_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_runs
    ADD CONSTRAINT workflow_runs_source_webhook_event_id_fkey FOREIGN KEY (source_webhook_event_id) REFERENCES public.webhook_events(webhook_event_id) ON DELETE SET NULL;


--
-- Name: workflow_runs workflow_runs_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_runs
    ADD CONSTRAINT workflow_runs_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES public.workflows(workflow_id) ON DELETE CASCADE;


--
-- Name: workflow_runs workflow_runs_workflow_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_runs
    ADD CONSTRAINT workflow_runs_workflow_version_id_fkey FOREIGN KEY (workflow_version_id) REFERENCES public.workflow_versions(workflow_version_id) ON DELETE RESTRICT;


--
-- Name: workflow_schedule_state workflow_schedule_state_last_workflow_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_schedule_state
    ADD CONSTRAINT workflow_schedule_state_last_workflow_run_id_fkey FOREIGN KEY (last_workflow_run_id) REFERENCES public.workflow_runs(workflow_run_id) ON DELETE SET NULL;


--
-- Name: workflow_schedule_state workflow_schedule_state_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_schedule_state
    ADD CONSTRAINT workflow_schedule_state_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES public.workflows(workflow_id) ON DELETE CASCADE;


--
-- Name: workflow_schedule_state workflow_schedule_state_workflow_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_schedule_state
    ADD CONSTRAINT workflow_schedule_state_workflow_version_id_fkey FOREIGN KEY (workflow_version_id) REFERENCES public.workflow_versions(workflow_version_id) ON DELETE RESTRICT;


--
-- Name: workflow_versions workflow_versions_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_versions
    ADD CONSTRAINT workflow_versions_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: workflow_versions workflow_versions_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_versions
    ADD CONSTRAINT workflow_versions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- Name: workflow_versions workflow_versions_restored_from_workflow_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_versions
    ADD CONSTRAINT workflow_versions_restored_from_workflow_version_id_fkey FOREIGN KEY (restored_from_workflow_version_id) REFERENCES public.workflow_versions(workflow_version_id) ON DELETE SET NULL;


--
-- Name: workflow_versions workflow_versions_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_versions
    ADD CONSTRAINT workflow_versions_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES public.workflows(workflow_id) ON DELETE CASCADE;


--
-- Name: workflows workflows_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT workflows_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;


--
-- Name: workflows workflows_current_workflow_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT workflows_current_workflow_version_id_fkey FOREIGN KEY (current_workflow_version_id) REFERENCES public.workflow_versions(workflow_version_id) ON DELETE SET NULL;


--
-- Name: workflows workflows_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT workflows_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--


--
-- PostgreSQL database dump
--


-- Dumped from database version 16.15 (Ubuntu 16.15-0ubuntu0.24.04.1)
-- Dumped by pg_dump version 16.15 (Ubuntu 16.15-0ubuntu0.24.04.1)

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
-- Data for Name: action_dependency_types; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.action_dependency_types VALUES ('5c7e240f-e923-4ade-ae34-67f13761dfbb', 'max.bot', 'MAX messenger bot', 'max', '{"type": "object", "required": ["bot_token"], "properties": {"bot_token": {"type": "string"}}}', '{"type": "object", "properties": {"default_format": {"enum": ["markdown", "html"]}, "default_notify": {"type": "boolean"}, "default_chat_id": {"type": "string"}}}', '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_dependency_types VALUES ('63caa881-fc7c-4bec-8d04-d02c835d0159', 'telegram.bot', 'Telegram bot', 'telegram', '{"type": "object", "required": ["bot_token"], "properties": {"bot_token": {"type": "string"}}}', '{"type": "object", "properties": {"parse_mode": {"enum": ["MarkdownV2", "HTML", "plain"]}, "default_chat_id": {"type": "string"}}}', '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_dependency_types VALUES ('d2daebd4-0984-4b85-8739-f68c030c1e7c', 'smtp.server', 'SMTP relay', 'email', '{"type": "object", "properties": {"password": {"type": "string"}}}', '{"type": "object", "properties": {"host": {"type": "string"}, "port": {"type": "integer"}, "username": {"type": "string"}, "from_name": {"type": "string"}, "encryption": {"enum": ["none", "starttls", "tls"]}, "from_email": {"type": "string"}}}', '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_dependency_types VALUES ('c5ca10b3-37df-419c-964e-09469845f4fa', 'email.template', 'Email HTML/text template binding', 'email', '{}', '{"type": "object", "properties": {"email_template_id": {"type": "string", "format": "uuid"}}}', '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_dependency_types VALUES ('8e388e6b-4297-43ee-aecb-01bf12fb9911', 'telegram.relay', 'Telegram relay', 'telegram', '{"type": "object", "required": ["bot_token"], "properties": {"bot_token": {"type": "string"}}}', '{"type": "object", "required": ["relay_key", "webhook_secret"], "properties": {"relay_key": {"type": "string"}, "relay_url": {"type": "string"}, "polling_limit": {"type": "integer", "maximum": 100, "minimum": 1}, "country_relays": {"type": "array"}, "webhook_secret": {"type": "string"}, "allowed_updates": {"type": "array", "items": {"type": "string"}}, "default_chat_id": {"type": "string"}, "polling_enabled": {"type": "boolean"}, "polling_delete_webhook": {"type": "boolean"}, "polling_timeout_seconds": {"type": "integer", "maximum": 60, "minimum": 1}, "polling_drop_pending_updates": {"type": "boolean"}}}', '2026-11-08 13:11:49.979983+03', '2026-11-08 13:11:50.039532+03');
INSERT INTO public.action_dependency_types VALUES ('275580da-4eef-48aa-87be-eaeaa31955b4', 'dadata.party', 'DaData party', 'dadata', '{"type": "object", "required": ["api_key"], "properties": {"api_key": {"type": "string"}}}', '{"type": "object", "properties": {"base_url": {"type": "string", "default": "https://suggestions.dadata.ru/suggestions/api/4_1/rs"}}}', '2026-11-08 13:11:50.073096+03', '2026-11-08 13:11:50.073096+03');
INSERT INTO public.action_dependency_types VALUES ('1e267e52-cf5a-41d4-851d-e1c194323c4f', 'ai.openai', 'OpenAI', 'ai.openai', '{"type": "object", "required": ["api_key"], "properties": {"api_key": {"type": "string"}}}', '{"type": "object", "properties": {"base_url": {"type": "string"}, "project_id": {"type": "string"}, "enable_tools": {"type": "boolean"}, "organization_id": {"type": "string"}, "privacy_profile": {"type": "string"}, "default_model_id": {"type": "string"}, "allowed_model_ids": {"type": "array", "items": {"type": "string"}}, "enable_patch_tool": {"type": "boolean"}, "model_refresh_ttl_seconds": {"type": "integer"}}}', '2026-11-08 13:11:50.716385+03', '2026-11-08 13:11:50.716385+03');
INSERT INTO public.action_dependency_types VALUES ('7d21cf43-c941-4a6d-9fd9-132eaa1b8df6', 'ai.openrouter', 'OpenRouter', 'ai.openrouter', '{"type": "object", "required": ["api_key"], "properties": {"api_key": {"type": "string"}}}', '{"type": "object", "properties": {"base_url": {"type": "string"}, "require_zdr": {"type": "boolean"}, "allow_fallbacks": {"type": "boolean"}, "default_model_id": {"type": "string"}, "allowed_model_ids": {"type": "array", "items": {"type": "string"}}}}', '2026-11-08 13:11:50.716385+03', '2026-11-08 13:11:50.716385+03');
INSERT INTO public.action_dependency_types VALUES ('c88ff5e0-8c80-4f47-8a2e-2208b196b973', 'ai.google_gemini', 'Google Gemini', 'ai.google_gemini', '{"type": "object", "required": ["api_key"], "properties": {"api_key": {"type": "string"}}}', '{"type": "object", "properties": {"base_url": {"type": "string"}, "default_model_id": {"type": "string"}, "allowed_model_ids": {"type": "array", "items": {"type": "string"}}}}', '2026-11-08 13:11:50.716385+03', '2026-11-08 13:11:50.716385+03');
INSERT INTO public.action_dependency_types VALUES ('5e62bec3-37ae-45e4-8aaa-638416bd942f', 'ai.anthropic', 'Anthropic Claude', 'ai.anthropic', '{"type": "object", "required": ["api_key"], "properties": {"api_key": {"type": "string"}}}', '{"type": "object", "properties": {"base_url": {"type": "string"}, "default_model_id": {"type": "string"}, "allowed_model_ids": {"type": "array", "items": {"type": "string"}}}}', '2026-11-08 13:11:50.716385+03', '2026-11-08 13:11:50.716385+03');
INSERT INTO public.action_dependency_types VALUES ('29e3fb15-35b0-48b8-93c3-8dfdb2f992d0', 'ai.mistral', 'Mistral AI', 'ai.mistral', '{"type": "object", "required": ["api_key"], "properties": {"api_key": {"type": "string"}}}', '{"type": "object", "properties": {"base_url": {"type": "string"}, "default_model_id": {"type": "string"}, "allowed_model_ids": {"type": "array", "items": {"type": "string"}}}}', '2026-11-08 13:11:50.716385+03', '2026-11-08 13:11:50.716385+03');
INSERT INTO public.action_dependency_types VALUES ('916e593c-dc8b-4295-aa42-4a82aff5be9e', 'ai.groq', 'Groq', 'ai.groq', '{"type": "object", "required": ["api_key"], "properties": {"api_key": {"type": "string"}}}', '{"type": "object", "properties": {"base_url": {"type": "string"}, "default_model_id": {"type": "string"}, "allowed_model_ids": {"type": "array", "items": {"type": "string"}}}}', '2026-11-08 13:11:50.716385+03', '2026-11-08 13:11:50.716385+03');
INSERT INTO public.action_dependency_types VALUES ('4bb16524-229c-46d8-a348-ee53e03aa94f', 'ai.xai', 'xAI', 'ai.xai', '{"type": "object", "required": ["api_key"], "properties": {"api_key": {"type": "string"}}}', '{"type": "object", "properties": {"base_url": {"type": "string"}, "default_model_id": {"type": "string"}, "allowed_model_ids": {"type": "array", "items": {"type": "string"}}}}', '2026-11-08 13:11:50.716385+03', '2026-11-08 13:11:50.716385+03');
INSERT INTO public.action_dependency_types VALUES ('f5ac00dc-c6a5-4a74-9bd8-dbfa29b06028', 'network.vless_proxy', 'VLESS route for AI egress', 'network.vless_proxy', '{"type": "object", "required": ["vless_url"], "properties": {"vless_url": {"type": "string"}}}', '{"type": "object", "properties": {"status": {"type": "string"}, "fail_mode": {"enum": ["fail_closed", "fail_open"]}, "route_key": {"type": "string"}, "route_kind": {"type": "string"}, "allow_hosts": {"type": "array", "items": {"type": "string"}}}}', '2026-11-08 13:11:50.716385+03', '2026-11-08 13:11:50.716385+03');
INSERT INTO public.action_dependency_types VALUES ('b3a52492-f50d-4ea0-a6fb-719b26dfa82f', 'ai.apinet', 'APINET / Riders Agency AI Gateway', 'ai.apinet', '{"type": "object", "required": ["api_key"], "properties": {"api_key": {"type": "string"}}}', '{"type": "object", "required": ["base_url"], "properties": {"base_url": {"type": "string", "default": "https://apinet.cloud/v1"}, "privacy_profile": {"type": "string", "default": "standard"}, "default_model_id": {"type": "string"}, "network_route_id": {"type": "string"}, "allowed_model_ids": {"type": "array", "items": {"type": "string"}}, "enable_embeddings": {"type": "boolean", "default": false}, "enable_tool_call_probe": {"type": "boolean", "default": true}, "enable_chat_completions": {"type": "boolean", "default": true}, "model_refresh_ttl_seconds": {"type": "integer", "default": 86400}}}', '2026-11-08 13:11:50.770691+03', '2026-11-08 13:11:50.770691+03');
INSERT INTO public.action_dependency_types VALUES ('17b74269-5f6a-495e-a98f-55c9fcf6b11e', 'uon.account', 'U-ON.Travel CRM', 'uon', '{"type": "object", "required": ["api_key"], "properties": {"api_key": {"type": "string"}}}', '{"type": "object", "properties": {"api_base": {"type": "string", "default": "https://api.u-on.ru"}, "countries": {"type": "string"}, "travel_type_id": {"type": "string"}, "allow_note_only": {"type": "boolean", "default": false}, "default_channel": {"type": "string"}, "extended_field_ids": {"type": "object", "properties": {"region": {"type": "string"}, "tour_name": {"type": "string"}, "tourist_count_text": {"type": "string"}}}, "lead_source_prefix": {"type": "string"}, "requirements_countries": {"type": "string"}}}', '2026-11-08 13:11:50.914039+03', '2026-11-08 13:11:50.914039+03');
INSERT INTO public.action_dependency_types VALUES ('3af16690-cb4c-4831-a894-014a17c4f3af', 'yclients.account', 'YClients · сертификаты', 'yclients', '{"type": "object", "required": ["partner_token", "receipt_hmac_secret"], "properties": {"login": {"type": "string"}, "password": {"type": "string"}, "user_token": {"type": "string"}, "partner_token": {"type": "string"}, "receipt_hmac_secret": {"type": "string"}}}', '{"type": "object", "properties": {"mode": {"enum": ["disabled", "verify", "write"], "type": "string"}, "api_base": {"type": "string", "default": "https://api.yclients.com/api/v1"}, "chain_id": {"type": "integer"}, "salon_ids": {"type": "array", "items": {"type": "integer"}}, "storage_id": {"type": "integer"}, "location_id": {"type": "integer"}, "receipt_url": {"type": "string"}, "receipt_audience": {"type": "string"}, "allow_type_create": {"type": "boolean", "default": false}, "receipt_allowed_hosts": {"type": "array", "items": {"type": "string"}}, "resolution_ttl_seconds": {"type": "integer", "default": 600}}}', '2026-11-08 13:11:51.117355+03', '2026-11-08 13:11:51.117355+03');


--
-- Data for Name: action_types; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.action_types VALUES ('d4dcac08-69d5-47c9-8471-0cf55f59bb23', 'webhook.trigger', 'Webhook trigger', 'trigger', '{}', '{}', false, NULL, '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_types VALUES ('0eb8c745-0baa-49d0-b272-a1dfd9f5e22b', 'manual.trigger', 'Manual trigger', 'trigger', '{}', '{}', false, NULL, '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_types VALUES ('501e4da3-387b-48df-be5c-4ef6ff2aeec5', 'condition.branch', 'Condition branch', 'logic', '{}', '{}', false, NULL, '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_types VALUES ('3817fdc8-1b9c-486b-90c2-fd96c43c0ee8', 'delay.wait', 'Delay', 'control', '{"type": "object", "properties": {"seconds": {"type": "integer", "minimum": 0}}}', '{}', false, NULL, '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_types VALUES ('7a12a46b-2fa8-4d6a-8490-344ae27d1dc0', 'http.request', 'HTTP request', 'integration', '{}', '{}', false, NULL, '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_types VALUES ('3bd4ec56-b9ad-4bd1-9c71-9d637b8bb22e', 'max.send_message', 'MAX: send message', 'communication', '{}', '{}', true, 'max', '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_types VALUES ('3d98ad18-472a-41e3-9d87-f4560558df78', 'telegram.send_message', 'Telegram: send message', 'communication', '{}', '{}', true, 'telegram', '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_types VALUES ('c7633ac4-57ca-4d25-b52e-2105124e842f', 'email.send', 'Email: send', 'communication', '{}', '{}', true, 'email', '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_types VALUES ('9c917740-4acb-4ff0-abef-f5e79ed6d05e', 'stop', 'Stop workflow', 'control', '{}', '{}', false, NULL, '2026-11-08 13:11:49.909729+03', '2026-11-08 13:11:49.909729+03');
INSERT INTO public.action_types VALUES ('0ae9b843-af16-408e-b8ba-9c00bc7f20a0', 'schedule.trigger', 'Scheduled trigger', 'trigger', '{"type": "object", "properties": {"enabled": {"type": "boolean"}, "timezone": {"type": "string"}, "schedule_key": {"type": "string"}, "interval_seconds": {"type": "integer", "minimum": 10}}}', '{}', false, NULL, '2026-11-08 13:11:49.937047+03', '2026-11-08 13:11:49.937047+03');
INSERT INTO public.action_types VALUES ('0941ff81-b40e-4d2b-b7ed-c5c0a1a4acdf', 'telegram.relay.trigger', 'Telegram relay trigger', 'trigger', '{"type": "object", "required": ["action_dependency_id"], "properties": {"relay_key": {"type": "string"}, "allowed_updates": {"type": "array", "items": {"type": "string"}}, "action_dependency_id": {"type": "string", "format": "uuid"}}}', '{"type": "object", "properties": {"text": {"type": "string"}, "chat_id": {"type": "string"}, "command": {"type": "string"}, "user_id": {"type": "string"}, "update_id": {"type": "integer"}, "message_id": {"type": "integer"}, "update_type": {"type": "string"}}}', true, 'telegram', '2026-11-08 13:11:49.979983+03', '2026-11-08 13:11:49.979983+03');
INSERT INTO public.action_types VALUES ('b61a6bbf-2b04-481e-86f2-cfe0574fbddf', 'bot.set_state', 'Bot: set state', 'control', '{"type": "object", "required": ["state_key"], "properties": {"state_key": {"type": "string"}, "context_json": {"type": "string"}, "merge_context": {"type": "boolean"}}}', '{"type": "object", "properties": {"state_key": {"type": "string"}, "context_json": {"type": "object"}, "bot_session_id": {"type": "string", "format": "uuid"}}}', false, 'telegram', '2026-11-08 13:11:50.008056+03', '2026-11-08 13:11:50.008056+03');
INSERT INTO public.action_types VALUES ('bd2ac5e3-8bb1-44a8-8f14-341143a39958', 'clients.payment_due.trigger', 'Client payment due trigger', 'trigger', '{"type": "object", "properties": {"enabled": {"type": "boolean"}, "timezone": {"type": "string", "default": "Europe/Moscow"}, "days_before_due": {"type": "integer", "default": 7, "maximum": 90, "minimum": 0}, "remind_every_days": {"type": "integer", "default": 7, "minimum": 1}}}', '{"type": "object", "properties": {"client": {"type": "object"}, "totals": {"type": "object"}, "project": {"type": "object"}, "services": {"type": "array"}, "payment_action_id": {"type": "string"}}}', false, NULL, '2026-11-08 13:11:50.073096+03', '2026-11-08 13:11:50.073096+03');
INSERT INTO public.action_types VALUES ('b5a85ee5-5820-41b1-895b-52bb1dc3ebc3', 'clients.payment_cycle.collect', 'Собрать цикл оплат', 'integration', '{"type": "object", "properties": {"days_before_due": {"type": "integer", "default": 30, "maximum": 90, "minimum": 0}}}', '{"type": "object", "properties": {"sent": {"type": "array"}, "counts": {"type": "object"}, "payments": {"type": "array"}, "scheduled": {"type": "array"}}}', false, NULL, '2026-11-08 13:11:50.196311+03', '2026-11-08 13:11:50.196311+03');
INSERT INTO public.action_types VALUES ('625ba58c-f11b-41e5-9fa8-495069327215', 'clients.manual_payment.trigger', 'Manual payment trigger', 'trigger', '{"type": "object", "properties": {}}', '{"type": "object", "properties": {"payment_event": {"type": "object"}, "manual_payment": {"type": "boolean"}}}', false, NULL, '2026-11-08 13:11:50.270334+03', '2026-11-08 13:11:50.270334+03');
INSERT INTO public.action_types VALUES ('35fcce69-b1e3-4e1c-a0b3-36d5855a89d0', 'site_chat.visitor_first_seen.trigger', 'Чат на сайте: первый заход посетителя', 'trigger', '{}', '{"type": "object", "properties": {"call_url": {"type": "string"}, "panel_url": {"type": "string"}, "visitor_id": {"type": "string", "format": "uuid"}, "current_url": {"type": "string"}, "current_title": {"type": "string"}, "chat_widget_id": {"type": "string", "format": "uuid"}, "conversation_id": {"type": "string", "format": "uuid"}, "conversation_url": {"type": "string"}, "visit_session_id": {"type": "string", "format": "uuid"}, "visitor_state_snapshot": {"type": "object"}}}', false, NULL, '2026-11-08 13:11:50.329974+03', '2026-11-08 13:11:50.386428+03');
INSERT INTO public.action_types VALUES ('61b17797-80fb-4360-b372-34888958c419', 'task.event_trigger', 'Событие задачи', 'trigger', '{"type": "object", "properties": {"event_key": {"enum": ["task.created", "task.updated", "task.status_changed", "task.comment_added", "time.logged"], "type": "string"}, "task_workflow_id": {"type": "string"}, "client_visible_only": {"type": "boolean", "default": false}}}', '{"type": "object", "properties": {"task": {"type": "object"}, "event": {"type": "object"}, "client": {"type": "object"}, "project": {"type": "object"}}}', false, NULL, '2026-11-08 13:11:50.668949+03', '2026-11-08 13:11:50.668949+03');
INSERT INTO public.action_types VALUES ('79eaba0a-da2a-47c8-94f1-1c2c70375202', 'telegram.unpin_message', 'Telegram: unpin message', 'communication', '{"type": "object", "required": ["action_dependency_id", "chat_id", "message_id"], "properties": {"chat_id": {"type": "string"}, "message_id": {"type": "string"}, "action_dependency_id": {"type": "string", "format": "uuid"}}}', '{"type": "object", "properties": {"unpinned": {"type": "boolean"}, "message_id": {"type": "integer"}}}', true, 'telegram', '2026-11-08 13:11:51.054953+03', '2026-11-08 13:11:51.054953+03');
INSERT INTO public.action_types VALUES ('aec45518-00d8-4d36-8e49-70f7dc5e5bc4', 'ai.agent.run', 'AI: запуск агента', 'integration', '{"type": "object", "required": ["integration", "model", "prompt"], "properties": {"model": {"type": "string"}, "prompt": {"type": "string"}, "timeout": {"type": "integer", "minimum": 1000}, "tool_mode": {"type": "string"}, "max_tokens": {"type": "integer", "minimum": 1}, "integration": {"type": "object", "required": ["action_dependency_id"], "properties": {"provider_key": {"type": "string"}, "action_dependency_id": {"type": "string"}}}, "json_schema": {"type": ["object", "null"]}, "temperature": {"type": "number", "maximum": 2, "minimum": 0}, "instructions": {"type": "string"}, "retry_policy": {"type": "object"}, "vpn_route_id": {"type": ["string", "null"]}, "response_mode": {"enum": ["text", "markdown", "json_schema", "tool_calls", "patch", "auto"], "type": "string"}, "tools_enabled": {"type": "boolean"}, "allow_background": {"type": "boolean"}, "legacy_output_compat": {"type": "boolean"}}}', '{"type": "object", "properties": {"model": {"type": "string"}, "usage": {"type": "object"}, "status": {"type": "string"}, "patches": {"type": "array"}, "provider": {"type": "string"}, "warnings": {"type": "array"}, "tool_calls": {"type": "array"}, "output_json": {"type": ["object", "array", "null"]}, "output_text": {"type": "string"}}}', true, 'ai', '2026-11-08 13:11:50.716385+03', '2026-11-08 13:11:50.716385+03');
INSERT INTO public.action_types VALUES ('d907103c-5f1c-4cf3-bf1a-dc849ddb93a0', 'person.fio.normalize', 'ФИО: нормализация', 'logic', '{"type": "object", "properties": {"input": {"type": "string", "description": "Шаблон ФИО, напр. {{ ag-backcall.name }}"}, "output_fields": {"type": "object", "additionalProperties": {"type": "string"}}}}', '{"type": "object", "properties": {"mode": {"type": "string"}, "u_name": {"type": "string"}, "changed": {"type": "boolean"}, "comment": {"type": "string"}, "u_sname": {"type": "string"}, "full_name": {"type": "string"}, "u_surname": {"type": "string"}}}', false, 'internal', '2026-11-08 13:11:50.919577+03', '2026-11-08 13:11:50.919577+03');
INSERT INTO public.action_types VALUES ('0afa5a72-a3e6-40ac-8a2e-d42c069a5c0a', 'telegram.send_document', 'Telegram: отправить документ', 'communication', '{"type": "object", "required": ["chat_id", "document_url"], "properties": {"caption": {"type": "string"}, "chat_id": {"type": "string"}, "document_url": {"type": "string"}, "retry_policy": {"type": "object"}, "action_dependency_id": {"type": "string", "format": "uuid"}, "disable_notification": {"type": "boolean"}}}', '{"type": "object", "properties": {"chat_id": {"type": "string"}, "delivered": {"type": "boolean"}, "message_id": {"type": "integer"}}}', true, 'telegram', '2026-11-08 13:11:50.919577+03', '2026-11-08 13:11:50.919577+03');
INSERT INTO public.action_types VALUES ('710645f1-ab3c-46b9-ad70-83b78484e02a', 'uon.create_lead', 'U-ON: создать обращение', 'integration', '{"type": "object", "properties": {"note": {"type": "string"}, "u_name": {"type": "string"}, "channel": {"type": "string"}, "u_email": {"type": "string"}, "u_phone": {"type": "string"}, "u_sname": {"type": "string"}, "u_surname": {"type": "string"}, "retry_policy": {"type": "object"}, "extra_fields_json": {"type": "string"}, "action_dependency_id": {"type": "string", "format": "uuid"}}}', '{"type": "object", "properties": {"note": {"type": "string"}, "result": {"type": "integer"}, "source": {"type": "string"}, "delivered": {"type": "boolean"}}}', true, 'uon', '2026-11-08 13:11:50.914039+03', '2026-11-08 13:11:50.919577+03');
INSERT INTO public.action_types VALUES ('72309b3b-4d15-4352-ae76-99832a34af54', 'person.phone.normalize', 'Телефон: нормализация', 'logic', '{"type": "object", "properties": {"input": {"type": "string", "description": "Шаблон телефона, напр. {{ ag-tour-claim.PHONE }}"}, "output_fields": {"type": "object", "additionalProperties": {"type": "string"}}}}', '{"type": "object", "properties": {"mode": {"type": "string"}, "changed": {"type": "boolean"}, "phone_raw": {"type": "string"}, "phone_e164": {"type": "string"}}}', false, 'internal', '2026-11-08 13:11:50.942963+03', '2026-11-08 13:11:50.942963+03');
INSERT INTO public.action_types VALUES ('249ccc62-54a1-4b77-9543-64eb00ec267b', 'ai.apinet.normalize_text', 'AI: normalize text with APINET', 'integration', '{"type": "object", "required": ["action_dependency_id", "source_text"], "properties": {"model": {"type": "string"}, "max_tokens": {"type": "integer", "minimum": 1}, "source_text": {"type": "string"}, "temperature": {"type": "number"}, "instructions": {"type": "string"}, "action_dependency_id": {"type": "string", "format": "uuid"}}}', '{"type": "object", "properties": {"model": {"type": "string"}, "usage": {"type": "object"}, "changed": {"type": "boolean"}, "original_text": {"type": "string"}, "normalized_text": {"type": "string"}}}', true, 'ai.apinet', '2026-11-08 13:11:51.04317+03', '2026-11-08 13:11:51.04317+03');
INSERT INTO public.action_types VALUES ('67ded9a1-8dbd-4171-adb1-8aebdae14e9d', 'telegram.edit_message_text', 'Telegram: edit message text', 'communication', '{"type": "object", "required": ["action_dependency_id", "chat_id", "message_id", "text"], "properties": {"text": {"type": "string"}, "chat_id": {"type": "string"}, "message_id": {"type": "string"}, "parse_mode": {"type": "string"}, "reply_markup_json": {"type": "string"}, "action_dependency_id": {"type": "string", "format": "uuid"}}}', '{"type": "object", "properties": {"edited": {"type": "boolean"}, "message_id": {"type": "integer"}}}', true, 'telegram', '2026-11-08 13:11:51.048894+03', '2026-11-08 13:11:51.048894+03');
INSERT INTO public.action_types VALUES ('0d9746ab-918f-4d49-af90-89567074b72e', 'telegram.answer_callback_query', 'Telegram: answer callback query', 'communication', '{"type": "object", "required": ["action_dependency_id", "callback_query_id"], "properties": {"text": {"type": "string"}, "show_alert": {"type": "boolean"}, "callback_query_id": {"type": "string"}, "action_dependency_id": {"type": "string", "format": "uuid"}}}', '{"type": "object", "properties": {"answered": {"type": "boolean"}}}', true, 'telegram', '2026-11-08 13:11:51.048894+03', '2026-11-08 13:11:51.048894+03');
INSERT INTO public.action_types VALUES ('6e336f30-2ff6-4fa6-868d-ccee4db37859', 'telegram.pin_message', 'Telegram: pin message', 'communication', '{"type": "object", "required": ["action_dependency_id", "chat_id", "message_id"], "properties": {"chat_id": {"type": "string"}, "message_id": {"type": "string"}, "action_dependency_id": {"type": "string", "format": "uuid"}, "disable_notification": {"type": "boolean"}}}', '{"type": "object", "properties": {"pinned": {"type": "boolean"}, "message_id": {"type": "integer"}}}', true, 'telegram', '2026-11-08 13:11:51.054953+03', '2026-11-08 13:11:51.054953+03');
INSERT INTO public.action_types VALUES ('b932bc3f-6404-4032-8a0b-9bb27aec864a', 'yclients.certificate.issue', 'YClients: выпустить сертификат', 'integration', '{"type": "object", "required": ["action_dependency_id", "operation_key", "order_id", "order_item_id", "instance_id", "certificate_number", "product_kind", "amount_rub"], "properties": {"order_id": {"type": "string", "format": "uuid"}, "amount_rub": {"type": "string"}, "instance_id": {"type": "string", "format": "uuid"}, "product_kind": {"enum": ["nominal", "service"], "type": "string"}, "retry_policy": {"type": "object"}, "operation_key": {"type": "string"}, "order_item_id": {"type": "string", "format": "uuid"}, "product_title": {"type": "string"}, "service_ids_json": {"type": "string"}, "certificate_number": {"type": "string"}, "action_dependency_id": {"type": "string", "format": "uuid"}}}', '{"type": "object", "properties": {"delivered": {"type": "boolean"}, "operation_key": {"type": "string"}, "transaction_id": {"type": "integer"}, "certificate_code": {"type": "string"}}}', true, 'yclients', '2026-11-08 13:11:51.117355+03', '2026-11-08 13:11:51.117355+03');


--
-- PostgreSQL database dump complete
--
