var menu_data = [
  { id:"authors", icon: "", value: "Authors", hotkey: "enter+shift"},
  { id:"books", icon: "", value:"Books" },
  { id:"genres", icon: "", value:"Genres" },
  { id:"series", icon:"", value:"Series" },
  { id:"contact", icon:"", value:"Contact Us" }
];

var header = {view:"toolbar",elements:[{ view:"label",  label: "mOPDS Simple Client"},{ view:"text", id:"grouplist_input", placeholder:"Поиск ..."}]};
var authors= { id:"authors", css:"preview-box", rows:[
  {
    id:"authors_list",
    view:"datatable",
    columns:[
      { id:"id", header:"", css:{"text-align":"center"}, width:50, hidden: true},
      { id:"full_name",	header:"Name", 	width:200, fillspace: 4 },
    ],
    datafetch: 20,
    loadahead: 20,
    url:"/api/v1/authors",
      onDataRequest: function (start, count) {
        onBeforeFilterCount++;
        webix.ajax().get("/api/v1/authors?page="+onBeforeFilterCount+"&per_page="+count).then(function(data){
          console.log(data.json());
          this.clearAll();
          this.parse(data.json());
        });
        return false;
      }
  }
]};

var books = { id:"books", css:"preview-box", rows:[
  {
    id: "books_list",
    view : "dataview",
    select : true,
    scrollX : true,
    type:{
        height: 90,
        template:"<div class='overall'><div class='rank'>#filename#.</div><div class='title'>#title#</div><div class='year'>#docdate# year</div></div>"
    },
    on:{
      "onItemClick":function(id, e, trg){
        //webix.message("Click on row: " + id.row+", column: " + id.column)
        webix.message(id);
      },
      "onDataRequest": function (start, count) {
        onBeforeFilterCount++;
        webix.ajax().get("/api/v1/books?no-details=true&page="+onBeforeFilterCount+"&per_page="+count).then(function(data){
          console.log(data.json());
          this.parse(data.json());
        });
        return false;
      }
    },
    url: "/api/v1/books?no-details=true"
  }
]};

var genres = { id:"genres", css:"preview-box", rows: [
  {
      height: 35,
      view:"toolbar",
      elements:[
          {view:"text", id:"list_genre_input",label:"Filter list by a genre",css:"fltr", labelWidth:170}
      ]
  },
  {
      view:"grouplist",
      id:"genres_list",
      filterMode:{
        level:0,
          showSubItems:1
      },
      autoConfig: true,
      select:true,
      on:{
        "onItemClick":function(id, e, trg){
          var genre = this.getItem(id);
          webix.message(genre.value + ' ' + genre.genre_id);
        }
      },
  }
]};

var series = { id:"series", css:"preview-box", rows: [
  {
      height: 35,
      view:"toolbar",
      elements:[
          {view:"text", id:"list_serie_input",label:"Filter list by a serie",css:"fltr", labelWidth:170}
      ]
  },
  {
    id:"series_list",
    view:"list",
    template:"#ser#",
    select:true,
    on:{
      "onItemClick":function(id, e, trg){
          webix.message(id);
      }
    },
    scheme:{
      $sort:{
        by:"ser",
        dir: 'asc'
      }
    },
    uniteBy:function(obj){
      return obj.full_name.substr(0,1);
    },
  }
]};

var contacts = { id:"contact", css:"preview-box", rows: [
  {
    id: "contact_list",
    view : "datatable",
    columns : [
      { id : "ID", header : "", width : 40, sort : "string", hidden: true },
      { id : "author", header : "Title", fillspace : true, sort : "string" },
      { id : "email", header : "FileName", fillspace : true, sort : "string" },
      { id : "project.name", header : "Size", fillspace : true, sort : "string" },
      { id : "project.version", header : "Size", fillspace : true, sort : "string" },
      { id : "project.link", header : "Size", fillspace : true, sort : "string" },
      { id : "project.created", header : "Size", fillspace : true, sort : "string" },
    ],
    url: "/api/v1/about",
    select : true,
    scrollX : true,
    datatype : "json",
    autoConfig: true
  }
]};

var onBeforeFilterCount = 0;
webix.ready(function(){
    if (webix.CustomScroll && !webix.env.touch)
        webix.CustomScroll.init();

    webix.ui({
      rows:[
        header,
        { animate:{type:"slide",subtype:"out", direction:"bottom"}, cells:[
          authors,
          books,
          genres,
          series,
          contacts
        ]},
        { view:"tabbar", type:"bottom", id:"tab", height:40, options:menu_data, multiview:true }
      ]
    });

    webix.ui.fullScreen();

    webix.ajax().get("/api/v1/books?no-details=true", function(text, data, XmlHttpRequest){
      $$("books_list").parse(data.json().items, "json");
    });

    webix.ajax().get("/api/v1/genres/menu", function(text, data, XmlHttpRequest){
      console.log(data.json());
      $$("genres_list").parse(data.json(), "json");
    });

    webix.ajax().get("/api/v1/series", function(text, data, XmlHttpRequest){
      $$("series_list").parse(data.json().items, "json");
    });
    $$("list_genre_input").attachEvent("onTimedKeyPress",function(){
        var value = this.getValue().toLowerCase();
        $$("genres_list").filter(function(obj){
            return obj.value.toLowerCase().indexOf(value)==0;
        });
    });
    $$("list_serie_input").attachEvent("onTimedKeyPress",function(){
        var value = this.getValue().toLowerCase();
        $$("series_list").filter(function(obj){
            return obj.ser.toLowerCase().indexOf(value)==0;
        });
    });
});
