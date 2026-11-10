//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//
// @dart=2.18

// ignore_for_file: unused_element, unused_import
// ignore_for_file: always_put_required_named_parameters_first
// ignore_for_file: constant_identifier_names
// ignore_for_file: lines_longer_than_80_chars

part of openapi.api;

class ProposalPatch {
  /// Returns a new [ProposalPatch] instance.
  ProposalPatch({
    this.inputJson = const {},
    this.editReason,
  });

  Map<String, Object> inputJson;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  String? editReason;

  @override
  bool operator ==(Object other) => identical(this, other) || other is ProposalPatch &&
    _deepEquality.equals(other.inputJson, inputJson) &&
    other.editReason == editReason;

  @override
  int get hashCode =>
    // ignore: unnecessary_parenthesis
    (inputJson.hashCode) +
    (editReason == null ? 0 : editReason!.hashCode);

  @override
  String toString() => 'ProposalPatch[inputJson=$inputJson, editReason=$editReason]';

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{};
      json[r'input_json'] = this.inputJson;
    if (this.editReason != null) {
      json[r'edit_reason'] = this.editReason;
    } else {
      json[r'edit_reason'] = null;
    }
    return json;
  }

  /// Returns a new [ProposalPatch] instance and imports its values from
  /// [value] if it's a [Map], null otherwise.
  // ignore: prefer_constructors_over_static_methods
  static ProposalPatch? fromJson(dynamic value) {
    if (value is Map) {
      final json = value.cast<String, dynamic>();

      // Ensure that the map contains the required keys.
      // Note 1: the values aren't checked for validity beyond being non-null.
      // Note 2: this code is stripped in release mode!
      assert(() {
        requiredKeys.forEach((key) {
          assert(json.containsKey(key), 'Required key "ProposalPatch[$key]" is missing from JSON.');
          assert(json[key] != null, 'Required key "ProposalPatch[$key]" has a null value in JSON.');
        });
        return true;
      }());

      return ProposalPatch(
        inputJson: mapCastOfType<String, Object>(json, r'input_json')!,
        editReason: mapValueOfType<String>(json, r'edit_reason'),
      );
    }
    return null;
  }

  static List<ProposalPatch> listFromJson(dynamic json, {bool growable = false,}) {
    final result = <ProposalPatch>[];
    if (json is List && json.isNotEmpty) {
      for (final row in json) {
        final value = ProposalPatch.fromJson(row);
        if (value != null) {
          result.add(value);
        }
      }
    }
    return result.toList(growable: growable);
  }

  static Map<String, ProposalPatch> mapFromJson(dynamic json) {
    final map = <String, ProposalPatch>{};
    if (json is Map && json.isNotEmpty) {
      json = json.cast<String, dynamic>(); // ignore: parameter_assignments
      for (final entry in json.entries) {
        final value = ProposalPatch.fromJson(entry.value);
        if (value != null) {
          map[entry.key] = value;
        }
      }
    }
    return map;
  }

  // maps a json object with a list of ProposalPatch-objects as value to a dart map
  static Map<String, List<ProposalPatch>> mapListFromJson(dynamic json, {bool growable = false,}) {
    final map = <String, List<ProposalPatch>>{};
    if (json is Map && json.isNotEmpty) {
      // ignore: parameter_assignments
      json = json.cast<String, dynamic>();
      for (final entry in json.entries) {
        map[entry.key] = ProposalPatch.listFromJson(entry.value, growable: growable,);
      }
    }
    return map;
  }

  /// The list of required keys that must be present in a JSON.
  static const requiredKeys = <String>{
    'input_json',
  };
}
