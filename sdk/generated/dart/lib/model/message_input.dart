//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//
// @dart=2.18

// ignore_for_file: unused_element, unused_import
// ignore_for_file: always_put_required_named_parameters_first
// ignore_for_file: constant_identifier_names
// ignore_for_file: lines_longer_than_80_chars

part of openapi.api;

class MessageInput {
  /// Returns a new [MessageInput] instance.
  MessageInput({
    required this.body,
    this.pageContextJson = const {},
    this.clientContextJson = const {},
    this.activeToolCallId,
    this.localDate,
    this.timezone,
  });

  String body;

  Map<String, Object> pageContextJson;

  Map<String, Object> clientContextJson;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  String? activeToolCallId;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  String? localDate;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  String? timezone;

  @override
  bool operator ==(Object other) => identical(this, other) || other is MessageInput &&
    other.body == body &&
    _deepEquality.equals(other.pageContextJson, pageContextJson) &&
    _deepEquality.equals(other.clientContextJson, clientContextJson) &&
    other.activeToolCallId == activeToolCallId &&
    other.localDate == localDate &&
    other.timezone == timezone;

  @override
  int get hashCode =>
    // ignore: unnecessary_parenthesis
    (body.hashCode) +
    (pageContextJson.hashCode) +
    (clientContextJson.hashCode) +
    (activeToolCallId == null ? 0 : activeToolCallId!.hashCode) +
    (localDate == null ? 0 : localDate!.hashCode) +
    (timezone == null ? 0 : timezone!.hashCode);

  @override
  String toString() => 'MessageInput[body=$body, pageContextJson=$pageContextJson, clientContextJson=$clientContextJson, activeToolCallId=$activeToolCallId, localDate=$localDate, timezone=$timezone]';

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{};
      json[r'body'] = this.body;
      json[r'page_context_json'] = this.pageContextJson;
      json[r'client_context_json'] = this.clientContextJson;
    if (this.activeToolCallId != null) {
      json[r'active_tool_call_id'] = this.activeToolCallId;
    } else {
      json[r'active_tool_call_id'] = null;
    }
    if (this.localDate != null) {
      json[r'local_date'] = this.localDate;
    } else {
      json[r'local_date'] = null;
    }
    if (this.timezone != null) {
      json[r'timezone'] = this.timezone;
    } else {
      json[r'timezone'] = null;
    }
    return json;
  }

  /// Returns a new [MessageInput] instance and imports its values from
  /// [value] if it's a [Map], null otherwise.
  // ignore: prefer_constructors_over_static_methods
  static MessageInput? fromJson(dynamic value) {
    if (value is Map) {
      final json = value.cast<String, dynamic>();

      // Ensure that the map contains the required keys.
      // Note 1: the values aren't checked for validity beyond being non-null.
      // Note 2: this code is stripped in release mode!
      assert(() {
        requiredKeys.forEach((key) {
          assert(json.containsKey(key), 'Required key "MessageInput[$key]" is missing from JSON.');
          assert(json[key] != null, 'Required key "MessageInput[$key]" has a null value in JSON.');
        });
        return true;
      }());

      return MessageInput(
        body: mapValueOfType<String>(json, r'body')!,
        pageContextJson: mapCastOfType<String, Object>(json, r'page_context_json') ?? const {},
        clientContextJson: mapCastOfType<String, Object>(json, r'client_context_json') ?? const {},
        activeToolCallId: mapValueOfType<String>(json, r'active_tool_call_id'),
        localDate: mapValueOfType<String>(json, r'local_date'),
        timezone: mapValueOfType<String>(json, r'timezone'),
      );
    }
    return null;
  }

  static List<MessageInput> listFromJson(dynamic json, {bool growable = false,}) {
    final result = <MessageInput>[];
    if (json is List && json.isNotEmpty) {
      for (final row in json) {
        final value = MessageInput.fromJson(row);
        if (value != null) {
          result.add(value);
        }
      }
    }
    return result.toList(growable: growable);
  }

  static Map<String, MessageInput> mapFromJson(dynamic json) {
    final map = <String, MessageInput>{};
    if (json is Map && json.isNotEmpty) {
      json = json.cast<String, dynamic>(); // ignore: parameter_assignments
      for (final entry in json.entries) {
        final value = MessageInput.fromJson(entry.value);
        if (value != null) {
          map[entry.key] = value;
        }
      }
    }
    return map;
  }

  // maps a json object with a list of MessageInput-objects as value to a dart map
  static Map<String, List<MessageInput>> mapListFromJson(dynamic json, {bool growable = false,}) {
    final map = <String, List<MessageInput>>{};
    if (json is Map && json.isNotEmpty) {
      // ignore: parameter_assignments
      json = json.cast<String, dynamic>();
      for (final entry in json.entries) {
        map[entry.key] = MessageInput.listFromJson(entry.value, growable: growable,);
      }
    }
    return map;
  }

  /// The list of required keys that must be present in a JSON.
  static const requiredKeys = <String>{
    'body',
  };
}
