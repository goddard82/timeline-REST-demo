from flask import Flask, jsonify, request, make_response, abort
from flask_httpauth import HTTPBasicAuth
from flask_json_schema import JsonSchema, JsonValidationError
from psycopg2.extras import Json, RealDictCursor
import psycopg2
import os
import datetime
import json
import uuid


DB_CREDS = json.loads(os.environ.get('DB_CREDS'))


app = Flask(__name__)

schema = JsonSchema(app)

event_schema = {
    'required': ['subsystem', 'type', 'status'],
    'properties': {
        'description': {'type': 'string'},
        'subsystem': {'type': 'string'},
        'type': {'type': 'string'},
        'status': {'type': 'string'},
        'payload': {'type': 'object'}
    }
}

auth = HTTPBasicAuth()

conn = None


def get_db():
    global conn
    if conn is None:
        conn = psycopg2.connect(host=DB_CREDS['hostname'], database=DB_CREDS['database'],
                                user=DB_CREDS['username'],
                                password=DB_CREDS['password'])
        conn.autocommit = True
    return conn


@auth.get_password
def get_password(username):
    if username == 'timeline':
        return 'p@S$wOrd'
    return None


@app.errorhandler(JsonValidationError)
def validation_error(e):
    return jsonify({'error': e.message, 'errors': [validation_error.message for validation_error in e.errors]}), 400


@app.errorhandler
def unauthorized():
    return make_response(jsonify({'error': 'Unauthorized access'}), 403)


@app.errorhandler(404)
def not_found(error):
    return make_response(jsonify({'error': 'Not found'}), 404)


@app.route('/timeline/events', methods=['GET'])
@auth.login_required
def get_all_events():
    global output
    upto = request.args.get('upto')
    limit = request.args.get('limit')
    offset = request.args.get('offset')
    if not limit:
        limit = None
    if not offset:
        offset = 0
    if not upto or upto != "now":
        sql = "SELECT * from timeline ORDER BY timestamp desc LIMIT %s OFFSET %s"
    if upto == "now":
        sql = "SELECT * from timeline where timestamp < now() ORDER BY timestamp desc LIMIT %s OFFSET %s"
    params = [limit, offset]
    try:
        conn = get_db()
        cur = conn.cursor(cursor_factory=RealDictCursor)
        cur.execute(sql, params)
        output = cur.fetchall()
        print(cur.mogrify(sql, params))
    except (Exception, psycopg2.DatabaseError, psycopg2.OperationalError, psycopg2.ProgrammingError,
            psycopg2.InterfaceError, psycopg2.DataError, psycopg2.InternalError) as error:
        print(error)
        conn.rollback()
    finally:
        if conn:
            cur.close()
        return jsonify({'events': output}), 200


@app.route('/timeline/events/recent', methods=['GET'])
@auth.login_required
def get_recent_events():
    global rows
    try:
        conn = get_db()
        cur = conn.cursor(cursor_factory=RealDictCursor)
        sql = "SELECT * from timeline WHERE timestamp BETWEEN now()::timestamp - (interval '30m') AND now()::timestamp ORDER BY timestamp desc"
        cur.execute(sql)
        # print(cur.mogrify(sql))
        rows = cur.fetchall()
    except (
            Exception, psycopg2.DatabaseError, psycopg2.OperationalError, psycopg2.ProgrammingError,
            psycopg2.InterfaceError,
            psycopg2.DataError, psycopg2.InternalError) as error:
        print(error)
        conn.rollback()
    finally:
        if conn:
            cur.close()
        if len(rows) >= 1:
            return jsonify({'events': rows}), 200
        else:
            return jsonify({}), 204


@app.route('/timeline/events/daterange', methods=['GET'])
@auth.login_required
def get_daterange():
    """Finds events where the timestamp/START is in the range specified."""
    global recents
    start = request.args.get('start')
    end = request.args.get('end')
    sql = "SELECT * from timeline WHERE timestamp BETWEEN %s AND %s ORDER BY timestamp desc;"
    params = [start, end]
    print(params)
    try:
        conn = get_db()
        cur = conn.cursor(cursor_factory=RealDictCursor)
        cur.execute(sql, params)
        recents = cur.fetchall()
        print(cur.mogrify(sql, params))
    except (Exception, psycopg2.DatabaseError, psycopg2.OperationalError, psycopg2.ProgrammingError,
            psycopg2.InterfaceError, psycopg2.DataError, psycopg2.InternalError) as error:
        print(error)
        conn.rollback()
    finally:
        if conn:
            cur.close()
        if len(recents) >= 1:
            return jsonify({'events': recents}), 200
        else:
            return jsonify({}), 204


@app.route('/timeline/events/daterange/advanced', methods=['GET'])
@auth.login_required
def get_daterange_advanced():
    global recents
    start = request.args.get('start') + str(' 00:00:00.000')
    end = request.args.get('end') + str(' 23:59:59.999')
    sql = "select * from timeline where (case when endtime isnull then timestamp between %s and %s else (tsrange(timestamp, endtime) && tsrange(%s, %s)) end) ORDER BY timestamp asc"
    params = [start, end, start, end]
    print(params)
    try:
        conn = get_db()
        cur = conn.cursor(cursor_factory=RealDictCursor)
        cur.execute(sql, params)
        recents = cur.fetchall()
        print(cur.mogrify(sql, params))
    except (Exception, psycopg2.DatabaseError, psycopg2.OperationalError, psycopg2.ProgrammingError,
            psycopg2.InterfaceError, psycopg2.DataError, psycopg2.InternalError) as error:
        print(error)
        conn.rollback()
    finally:
        if conn:
            cur.close()
        if len(recents) >= 1:
            return jsonify({'events': recents}), 200

        else:
            return jsonify({}), 204


@app.route('/timeline/events/query', methods=['GET'])
@auth.login_required
"""https://timeline.avkn.co/timeline/events/query?payload={
        "affectedServices": null,
        "message": "Users may have experienced intermittent issues with this service.",
        "usersAffected": "Some users were affected"
      }"""
def get_query():
    global recents
    long_dict = request.args
    print(long_dict)
    element_list = []
    for keys, values in long_dict.items():
        element_list.append({keys: values})
    sql = "SELECT * from timeline where "
    first_element = element_list[0]
    params = []
    for key, value in first_element.items():
        first_element_key = key
        params.append(value)
    print(first_element_key)
    sql += str(first_element_key) + " = %s "
    element_list.pop(0)
    for event_data_pair in element_list:
        for key, value in event_data_pair.items():
            if key == "description":
                sql += "AND description ILIKE %s "
                params.append(value)
            elif key == "payload":
                sql += "AND payload = %s "
                params.append(value)
            elif key == "start":
                sql += "AND " + str(key) + " = %s "
                params.append(value)
            else:
                sql += "AND " + str(key) + " = %s "
                params.append(value)
    sql += "order by timestamp desc"
    print(sql)
    print(params)
    try:
        conn = get_db()
        cur = conn.cursor(cursor_factory=RealDictCursor)
        cur.execute(sql, Json(params))
        recents = cur.fetchall()
        print(cur.mogrify(sql, params))
    except (Exception, psycopg2.DatabaseError, psycopg2.OperationalError, psycopg2.ProgrammingError,
            psycopg2.InterfaceError, psycopg2.DataError, psycopg2.InternalError) as error:
        print(error)
        conn.rollback()
    finally:
        if conn:
            cur.close()
        if len(recents) >= 1:
            return jsonify({'events': recents}), 200
        else:
            return jsonify({}), 204


@app.route('/timeline/events/', methods=['GET'])
@auth.login_required
def get_query_last_x():
    global recents
    last = request.args.get('last')
    sql = "WITH events AS (SELECT * from timeline WHERE timestamp <= now()::timestamp ORDER BY timestamp desc LIMIT %s) SELECT * FROM events ORDER BY timestamp desc;"
    try:
        conn = get_db()
        cur = conn.cursor(cursor_factory=RealDictCursor)
        cur.execute(sql, (last, ))
        recents = cur.fetchall()
        print(cur.mogrify(sql, [last]))
    except (Exception, psycopg2.DatabaseError, psycopg2.OperationalError, psycopg2.ProgrammingError, psycopg2.InterfaceError, psycopg2.DataError, psycopg2.InternalError) as error:
        print(error)
        conn.rollback()
    finally:
        if conn:
            cur.close()
        if len(recents) >= 1:
            return jsonify({'events': recents}), 200
        else:
            return jsonify({}), 204


@app.route('/timeline/events/<uuid:instance_id>', methods=['GET'])
@auth.login_required
def get_event(instance_id):
    psycopg2.extras.register_uuid(instance_id)
    global output
    sql = "SELECT * from timeline WHERE instance_id = %s"
    try:
        conn = get_db()
        cur = conn.cursor(cursor_factory=RealDictCursor)
        cur.execute(sql, [instance_id])
        output = cur.fetchall()
        # print(cur.mogrify(sql, (str(instance_id), output)))
    except (
            Exception, psycopg2.DatabaseError, psycopg2.OperationalError, psycopg2.ProgrammingError,
            psycopg2.InterfaceError,
            psycopg2.DataError, psycopg2.InternalError) as error:
        print(error)
        conn.rollback()
    finally:
        if conn:
            cur.close()
        if len(output) >= 1:
            return jsonify({'events': output}), 200
        else:
            return jsonify({}), 204


@app.route('/timeline/events', methods=['POST'])
@auth.login_required
@schema.validate(event_schema)
def create_event():
    instance_id = str(uuid.uuid4()),
    subsystem = request.json.get('subsystem').lower()
    event_type = request.json.get('type').lower()
    timestamp = request.json.get('timestamp')
    if not timestamp:
        timestamp = datetime.datetime.now()
    else:
        timestamp = request.json.get('timestamp')
    status = request.json.get('status').lower()
    description = request.json.get('description')
    payload = request.json.get('payload')
    endtime = request.json.get('endtime')
    pyload = Json(payload) if payload else None
    sql = """INSERT INTO timeline (instance_id, description, subsystem, type, status, timestamp, payload, endtime)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s)"""
    # suppressed = request.json.get('suppressed')
    try:
        conn = get_db()
        cur = conn.cursor(cursor_factory=RealDictCursor)
        cur.execute(sql, (instance_id, description, subsystem, event_type, status, timestamp, pyload, endtime))
        conn.commit()
        print(cur.mogrify(sql, (instance_id, description, subsystem, event_type, status, timestamp, pyload, endtime)))
        # conn.close()
    except (
            Exception, psycopg2.DatabaseError, psycopg2.OperationalError, psycopg2.ProgrammingError,
            psycopg2.InterfaceError,
            psycopg2.DataError, psycopg2.InternalError) as error:
        print(error)
        conn.rollback()
    finally:
        if conn:
            cur.close()
        if not request.json:
            exit(400)
        event = {
            'instance_id': instance_id[0],
            'subsystem': subsystem,
            'type': event_type,
            'timestamp': timestamp,
            'status': status,
            'description': description,
            'payload': payload,
            'endtime': endtime
            # 'suppressed': request.json.get('suppressed')
        }
        return jsonify({'events': event}), 201


@app.route('/timeline/events/airtable', methods=['POST'])
@auth.login_required
@schema.validate(event_schema)
def convert_airtable_event():
    """Creates an event but uses the existing timestamp as opposed to creating a new one. Used when scraping data from
    the temp airtable timeline"""
    instance_id = str(uuid.uuid4()),
    subsystem = request.json.get('subsystem').lower()
    event_type = request.json.get('type').lower()
    timestamp = request.json.get('timestamp')
    status = request.json.get('status').lower()
    description = request.json.get('description')
    payload = request.json.get('payload')
    endtime = request.json.get('endtime')
    sql = """INSERT INTO timeline (instance_id, description, subsystem, type, status, timestamp, payload, endtime)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s);
            """
    # suppressed = request.json.get('suppressed')
    try:
        conn = get_db()
        cur = conn.cursor(cursor_factory=RealDictCursor)
        cur.execute(sql, (instance_id, description, subsystem, event_type, status, timestamp, Json(payload), endtime))
        print(cur.mogrify(sql, (instance_id, description, subsystem, event_type, status, timestamp, Json(payload), endtime)))
        if not request.json:
            exit(400)
    except (Exception, psycopg2.DatabaseError, psycopg2.OperationalError, psycopg2.ProgrammingError,
            psycopg2.InterfaceError, psycopg2.DataError, psycopg2.InternalError) as error:
        print(error)
        conn.rollback()
    finally:
        if conn:
            cur.close()
        event = {
            'instance_id': instance_id,
            'subsystem': subsystem,
            'type': event_type,
            'timestamp': timestamp,
            'status': status,
            'description': description,
            'payload': payload,
            'endtime': endtime
            # 'suppressed': request.json.get('suppressed')
        }
        return jsonify({'events': event}), 201


@app.route('/timeline/events/<uuid:instance_id>', methods=['PUT'])
@auth.login_required
def update_event(instance_id):
    sql = """UPDATE timeline
    SET endtime = %s, status = 'ok'
    WHERE instance_id = %s"""
    endtime = datetime.datetime.now()
    try:
        conn = get_db()
        cur = conn.cursor(cursor_factory=RealDictCursor)
        psycopg2.extras.register_uuid(instance_id)
        cur.execute(sql, (endtime, instance_id))
        conn.commit()
        print(cur.mogrify(sql, (endtime, instance_id)))
    except (
            Exception, psycopg2.DatabaseError, psycopg2.OperationalError, psycopg2.ProgrammingError,
            psycopg2.InterfaceError,
            psycopg2.DataError, psycopg2.InternalError) as error:
        print(error)
        conn.rollback()
    finally:
        if conn:
            cur.close()
        return get_event(instance_id)


@app.route('/timeline/events/<uuid:instance_id>', methods=['DELETE'])
@auth.login_required
def delete_event(instance_id):
    sql = """DELETE from timeline
    WHERE instance_id = %s"""
    psycopg2.extras.register_uuid(instance_id)
    # rows_deleted = 0
    # deletee = str(instance_id)
    try:
        conn = get_db()
        cur = conn.cursor(cursor_factory=RealDictCursor)
        cur.execute(sql, (instance_id, ))
        rows_deleted = cur.rowcount
    except (
            Exception, psycopg2.DatabaseError, psycopg2.OperationalError, psycopg2.ProgrammingError,
            psycopg2.InterfaceError,
            psycopg2.DataError, psycopg2.InternalError) as error:
        print(error)
        conn.rollback()
    finally:
        if conn:
            cur.close()
        event = {
            'instance_id': instance_id
        }
        return jsonify({'deleted': event}), 200


@app.route('/timeline/health', methods=['GET'])
def health():
    return {'message': 'Healthy'}, 200


if __name__ == "__main__":
    app.run(host='0.0.0.0', debug=True)
