//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//
// @dart=2.18

// ignore_for_file: unused_element, unused_import
// ignore_for_file: always_put_required_named_parameters_first
// ignore_for_file: constant_identifier_names
// ignore_for_file: lines_longer_than_80_chars

part of openapi.api;

class TaskPage {
  /// Returns a new [TaskPage] instance.
  TaskPage({
    required this.tasks,
  });

  TaskList tasks;

  @override
  bool operator ==(Object other) => identical(this, other) || other is TaskPage &&
    other.tasks == tasks;

  @override
  int get hashCode =>
    // ignore: unnecessary_parenthesis
    (tasks.hashCode);

  @override
  String toString() => 'TaskPage[tasks=$tasks]';

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{};
      json[r'tasks'] = this.tasks;
    return json;
  }

  /// Returns a new [TaskPage] instance and imports its values from
  /// [value] if it's a [Map], null otherwise.
  // ignore: prefer_constructors_over_static_methods
  static TaskPage? fromJson(dynamic value) {
    if (value is Map) {
      final json = value.cast<String, dynamic>();

      // Ensure that the map contains the required keys.
      // Note 1: the values aren't checked for validity beyond being non-null.
      // Note 2: this code is stripped in release mode!
      assert(() {
        requiredKeys.forEach((key) {
          assert(json.containsKey(key), 'Required key "TaskPage[$key]" is missing from JSON.');
          assert(json[key] != null, 'Required key "TaskPage[$key]" has a null value in JSON.');
        });
        return true;
      }());

      return TaskPage(
        tasks: TaskList.fromJson(json[r'tasks'])!,
      );
    }
    return null;
  }

  static List<TaskPage> listFromJson(dynamic json, {bool growable = false,}) {
    final result = <TaskPage>[];
    if (json is List && json.isNotEmpty) {
      for (final row in json) {
        final value = TaskPage.fromJson(row);
        if (value != null) {
          result.add(value);
        }
      }
    }
    return result.toList(growable: growable);
  }

  static Map<String, TaskPage> mapFromJson(dynamic json) {
    final map = <String, TaskPage>{};
    if (json is Map && json.isNotEmpty) {
      json = json.cast<String, dynamic>(); // ignore: parameter_assignments
      for (final entry in json.entries) {
        final value = TaskPage.fromJson(entry.value);
        if (value != null) {
          map[entry.key] = value;
        }
      }
    }
    return map;
  }

  // maps a json object with a list of TaskPage-objects as value to a dart map
  static Map<String, List<TaskPage>> mapListFromJson(dynamic json, {bool growable = false,}) {
    final map = <String, List<TaskPage>>{};
    if (json is Map && json.isNotEmpty) {
      // ignore: parameter_assignments
      json = json.cast<String, dynamic>();
      for (final entry in json.entries) {
        map[entry.key] = TaskPage.listFromJson(entry.value, growable: growable,);
      }
    }
    return map;
  }

  /// The list of required keys that must be present in a JSON.
  static const requiredKeys = <String>{
    'tasks',
  };
}
