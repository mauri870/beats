---
mapped_pages:
  - https://www.elastic.co/guide/en/beats/devguide/current/event-fields-yml.html
---

# Defining field mappings [event-fields-yml]

You must define the fields used by your Beat, along with their mapping details, in `_meta/fields.yml`. After editing this file, run `make update`.

Define the field mappings in the `fields` array:

```yaml
- key: mybeat
  title: mybeat
  description: These are the fields used by mybeat.
  fields:
    - name: last_name <1>
      type: keyword <2>
      required: true <3>
      description: > <4>
        The last name.
    - name: first_name
      type: keyword
      required: true
      description: >
        The first name.
    - name: comment
      type: text
      required: false
      description: >
        Comment made by the user.
```

1. `name`: The field name
2. `type`: The field type. The value of `type` can be any datatype [available in {{es}}](elasticsearch://reference/elasticsearch/mapping-reference/field-data-types.md). If no value is specified, the default type is `keyword`.
3. `required`: Whether or not a field value is required
4. `description`: Some information about the field contents


## Mapping parameters [_mapping_parameters]

You can specify other mapping parameters for each field. See the [{{es}} Reference](elasticsearch://reference/elasticsearch/mapping-reference/mapping-parameters.md) for more details about each parameter.

`format`
:   Specify a custom date format used by the field.

`multi_fields`
:   For `text` or `keyword` fields, use `multi_fields` to define multi-field mappings.

`enabled`
:   Whether or not the field is enabled.

`analyzer`
:   Which analyzer to use when indexing.

`search_analyzer`
:   Which analyzer to use when searching.

`norms`
:   Applies to `text` and `keyword` fields. Default is `false`.

`dynamic`
:   Dynamic field control. Can be one of `true` (default), `false`, or `strict`.

`index`
:   Whether or not the field should be indexed.

`doc_values`
:   Whether or not the field should have doc values generated.

`copy_to`
:   Which field to copy the field value into.

`ignore_above`
:   {{es}} ignores (does not index) strings that are longer than the specified value. When this property value is missing or `0`, the `libbeat` default value of `1024` characters is used. If the value is `-1`, the {{es}} default value is used.

For example, you can use the `copy_to` mapping parameter to copy the `last_name` and `first_name` fields into the `full_name` field at index time:

```yaml
- key: mybeat
  title: mybeat
  description: These are the fields used by mybeat.
  fields:
    - name: last_name
      type: text
      required: true
      copy_to: full_name <1>
      description: >
        The last name.
    - name: first_name
      type: text
      required: true
      copy_to: full_name <2>
      description: >
        The first name.
    - name: full_name
      type: text
      required: false
      description: >
        The last_name and first_name combined into one field for easy searchability.
```

1. Copy the value of `last_name` into `full_name`
2. Copy the value of `first_name` into `full_name`


There are also some {{kib}}-specific properties, not detailed here. These are: `analyzed`, `count`, `searchable`, `aggregatable`, and `script`. {{kib}} parameters can also be described using `pattern`, `input_format`, `output_format`, `output_precision`, `label_template`, `url_template`, and `open_link_in_current_tab`.


## Defining text multi-fields [_defining_text_multi_fields]

There are various options that you can apply when using text fields. You can define a simple text field using the default analyzer without any other options, as in the example shown earlier.

To keep the original keyword value when using `text` mappings, for instance to use in aggregations or ordering, you can use a multi-field mapping:

```yaml
- key: mybeat
  title: mybeat
  description: These are the fields used by mybeat.
  fields:
    - name: city
      type: text
      multi_fields: <1>
        - name: keyword <2>
          type: keyword <3>
```

1. `multi_fields`: Define the `multi_fields` mapping parameter.
2. `name`: This is a conventional name for a multi-field. It can be anything (`raw` is another common option) but the convention is to use `keyword`.
3. `type`: Specify the `keyword` type to use the field in aggregations or to order documents.


For more information, see the [{{es}} documentation about multi-fields](elasticsearch://reference/elasticsearch/mapping-reference/multi-fields.md).


## Defining a text analyzer in-line [_defining_a_text_analyzer_in_line]

It is possible to define a new text analyzer or search analyzer in-line with the field definition in the field’s mapping parameters.

For example, you can define a new text analyzer that does not break hyphenated names:

```yaml
- key: mybeat
  title: mybeat
  description: These are the fields used by mybeat.
  fields:
    - name: last_name
      type: text
      required: true
      description: >
        The last name.
      analyzer:
        mybeat_hyphenated_name: <1>
          type: pattern <2>
          pattern: "[\\W&&[^-]]+" <3>
      search_analyzer:
        mybeat_hyphenated_name: <4>
          type: pattern
          pattern: "[\\W&&[^-]]+"
```

1. Use a newly defined text analyzer
2. Define the custome analyzer type
3. Specify the analyzer behaviour
4. Use the same analyzer for the search


The names of custom analyzers that are defined in-line may not be reused for a different text analyzer. If a text analyzer name is reused it is checked for matching existing instances of the analyzer. It is recommended that the analyzer name is prefixed with the beat name to avoid name clashes.

For more information, see [{{es}} documentation about defining custom text analyzers](docs-content://manage-data/data-store/text-analysis/create-custom-analyzer.md).


