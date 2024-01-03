resource "spacelift_saved_filter" "my_filter" {
  type      = "webhooks"
  name      = "filter for all xyz teams"
  is_public = true
  data = jsonencode({
    "key" : "activeFilters",
    "value" : jsonencode({
      "filters" : [
        [
          "name",
          {
            "key" : "name",
            "filterName" : "name",
            "type" : "STRING",
            "values" : [
              "team_xyz_*"
            ]
          }
        ]
      ],
      "sort" : {
        "direction" : "ASC",
        "option" : "space"
      },
      "text" : null,
      "order" : [
        {
          "name" : "enabled",
          "visible" : true
        },
        {
          "name" : "endpoint",
          "visible" : true
        },
        {
          "name" : "slug",
          "visible" : true
        },
        {
          "name" : "label",
          "visible" : true
        },
        {
          "name" : "name",
          "visible" : true
        },
        {
          "name" : "space",
          "visible" : true
        }
      ]
    })
  })
}