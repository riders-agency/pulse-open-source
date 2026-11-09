//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//
// @dart=2.18

// ignore_for_file: unused_element, unused_import
// ignore_for_file: always_put_required_named_parameters_first
// ignore_for_file: constant_identifier_names
// ignore_for_file: lines_longer_than_80_chars

part of openapi.api;

class SessionInput {
  /// Returns a new [SessionInput] instance.
  SessionInput({
    required this.screenScope,
    required this.screenEntityType,
    required this.screenEntityId,
    this.pageContextJson = const {},
  });

  String screenScope;

  String screenEntityType;

  String screenEntityId;

  Map<String, Object> pageContextJson;

  @override
  bool operator ==(Object other) => identical(this, other) || other is SessionInput &&
    other.screenScope == screenScope &&
    other.screenEntityType == screenEntityType &&
    other.screenEntityId == screenEntityId &&
    _deepEquality.equals(other.pageContextJson, pageContextJson);

  @override
  int get hashCode =>
    // ignore: unnecessary_parenthesis
    (screenScope.hashCode) +
    (screenEntityType.hashCode) +
    (screenEntityId.hashCode) +
    (pageContextJson.hashCode);

  @override
  String toString() => 'SessionInput[screenScope=$screenScope, screenEntityType=$screenEntityType, screenEntityId=$screenEntityId, pageContextJson=$pageContextJson]';

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{};
      json[r'screen_scope'] = this.screenScope;
      json[r'screen_entity_type'] = this.screenEntityType;
      json[r'screen_entity_id'] = this.screenEntityId;
      json[r'page_context_json'] = this.pageContextJson;
    return json;
  }

  /// Returns a new [SessionInput] instance and imports its values from
  /// [value] if it's a [Map], null otherwise.
  // ignore: prefer_constructors_over_static_methods
  static SessionInput? fromJson(dynamic value) {
    if (value is Map) {
      final json = value.cast<String, dynamic>();

      // Ensure that the map contains the required keys.
      // Note 1: the values aren't checked for validity beyond being non-null.
      // Note 2: this code is stripped in release mode!
      assert(() {
        requiredKeys.forEach((key) {
          assert(json.containsKey(key), 'Required key "SessionInput[$key]" is missing from JSON.');
          assert(json[key] != null, 'Required key "SessionInput[$key]" has a null value in JSON.');
        });
        return true;
      }());

      return SessionInput(
        screenScope: mapValueOfType<String>(json, r'screen_scope')!,
        screenEntityType: mapValueOfType<String>(json, r'screen_entity_type')!,
        screenEntityId: mapValueOfType<String>(json, r'screen_entity_id')!,
        pageContextJson: mapCastOfType<String, Object>(json, r'page_context_json') ?? const {},
      );
    }
    return null;
  }

  static List<SessionInput> listFromJson(dynamic json, {bool growable = false,}) {
    final result = <SessionInput>[];
    if (json is List && json.isNotEmpty) {
      for (final row in json) {
        final value = SessionInput.fromJson(row);
        if (value != null) {
          result.add(value);
        }
      }
    }
    return result.toList(growable: growable);
  }

  static Map<String, SessionInput> mapFromJson(dynamic json) {
    final map = <String, SessionInput>{};
    if (json is Map && json.isNotEmpty) {
      json = json.cast<String, dynamic>(); // ignore: parameter_assignments
      for (final entry in json.entries) {
        final value = SessionInput.fromJson(entry.value);
        if (value != null) {
          map[entry.key] = value;
        }
      }
    }
    return map;
  }

  // maps a json object with a list of SessionInput-objects as value to a dart map
  static Map<String, List<SessionInput>> mapListFromJson(dynamic json, {bool growable = false,}) {
    final map = <String, List<SessionInput>>{};
    if (json is Map && json.isNotEmpty) {
      // ignore: parameter_assignments
      json = json.cast<String, dynamic>();
      for (final entry in json.entries) {
        map[entry.key] = SessionInput.listFromJson(entry.value, growable: growable,);
      }
    }
    return map;
  }

  /// The list of required keys that must be present in a JSON.
  static const requiredKeys = <String>{
    'screen_scope',
    'screen_entity_type',
    'screen_entity_id',
  };
}
