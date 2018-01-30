# 1. Basics

jQuery is a factory function rather than a constructor

4种调用方式：

-   传入 css 选择符
-   传入 element, document or Window object
-   string of html text
-   pass a function to it, If you do this, the function you pass will be invoked when the document has been loaded and the DOM is ready to be ma- nipulated.
      jQuery(function() { // Invoked when document has loaded // All jQuery code goes here
      });

# 2. Element Getters and Setters

- getting and setting html attributes: attr()
- getting and setting css attributes: css()
- Getting and Setting CSS Classes: addClass() and removeClass() toggleClass() hasClass()
- Getting and Setting HTML Form Values: val()
- Getting and Setting Element Content: The text() and html() methods query and set the plain-text or HTML content of an element.
- Getting and Setting Element Geometry: offset() position() innerWidth outerWidth()  scrollLeft()
- G>etting and Setting Element Data: data() removeData()

# 3. Altering Document Structure
- Inserting and Replacing Elements: append() prepend() after() before() replaceWith()
- Copying Elements: clone();clone() does not normally copy event handlers
- Wrapping Elements: wrap(), wrapInner(), wrapAll()
- Deleting Elements: empty() removes all children (including text nodes) of each of the selected elements without
  altering the elements themselves. The remove() method, by contrast, removes the selected elements (and all of their
  con- tent) from the document. the unwrap() method performs element removal in a way that is opposite of the wrap() or wrapAll() method: it re- moves the parent of each selected element without affecting the selected elements or their siblings.

# 4. Events
- Simple Event Handler Registration

  ```
  $('p').click(function() {
      $(this).css('background-color', 'gray');
  })
  blur() focusin() change() focusout() click() keydown() dblclick() keypress() error() keyup() focus() load()
  mousedown() mouseenter() mouseleave() mousemove() mouseout() mouseover()
  mouseup() resize() scroll() select() submit() unload()
  ```

- jQuery Event Handlers : That is, returning false is the same as calling the preventDefault() and stopPropagation() methods of the Event object
- The jQuery Event Object:
- Advanced Event Handler Registration: bind() one()
    ```
    $('p').click(f) -> $('p').bind('click', f);
      // Bind f as a mouseover handler in namespace "myMod"
      $('a').bind('mouseover.myMod', f);
    $('a').hover(f, g) <==> $('a').bind({mouseenter:f, mouseleave:g})
    ```
- Deregistering Event Handlers: unbind() to prevent it from being triggered by future events.
    ```
    $('a').unbind("mouseover mouseout");
    $('#mybutton').unbind('click', myClickHandler);
    $('a').unbind({ // Remove specific event handlers
      mouseover: mouseoverHandler,
      mouseout: mouseoutHandler
    });
    ```
- Triggering Events: After invoking event handlers, trigger() (and the convenience methods that call it) perform
  whatever default action is asso- ciated with the triggered event (assuming that the event han- dlers didn’t return
  false or call preventDefault() on the event object.If you want to invoke event handlers without performing the default action, use triggerHandler(), which works just like trigger() except that it first calls the preventDefault() and cancelBubble() methods of the Event object.

    ```
    // Act as if the user clicked the Submit button
    $("#my_form").submit();
    $('#my_form').trigger('submmit');

    // Trigger button click handlers in namespace ns1
    $("button").trigger("click.ns1");

    // Add extra property to the event object when triggering
    $('#button1').trigger({type:'click', synthetic:true});
    $('#button1').click(function(e) {
      if (e.synthetic) {...};
    });
    ```

- Custom Events:
- Live Events: bind()方法不会对动态创建的元素起作用，这种问题称之为"live envents". To use live events, use the delegate() and undelegate() methods instead of bind() and unbind()<Paste>
    jQuery defines a method named live() that can also be used to register live events.
    To deregister live event handlers, use die() or undelegate().

    ``` html
    $(document).delegate("a", "mouseover", linkHa)
    // Remove all live handlers for mouseover on <a> tags
    $('a').die('mouseover');
    // Remove just one specific live handler
    $('a').die('mouseover', linkHandler);
    ```

# 5 Animated Effects

- Simple Effects: Query’s animations are queued by default
      - fadeIn(), fadeOut(), fadeTo()
      - show(), hide(), toggle()
      - slideDown(), slideUp(), slideToggle()
- Custom Animations: animate()
- The Animation Properties Object

      ```
      $("p").animate({
        "margin-left": "+=.5in", // Increase indent
         opacity: "-=.1" // And decrease opacity
      });
      ```

- The Animation Options Object
- Canceling, Delaying, and Queuing Effects: stop(), delay()

# 6 Ajax
- The load() method: pass it a URL, which it will asynchronously load the content of, and then insert that content into each of the selected elements
- Ajax Utility Functions:
      - jQuery.getScript(): The jQuery.getScript() function takes the URL of a file of JavaScript code as its first argument. It asynchronously loads and then executes that code in the global scope. It can work for both same-origin and cross-origin scripts:
      - jQuery.getJSON():  it fetches text and then processes it specially before invoking the specified call- back. Instead of executing the text as a script, jQuery.getJSON() parses it as JSON (using the jQuery.parseJSON() function;

      ```
      // Suppose data.json contains the text: '{"x":1,"y":2}'
      jQuery.getJSON("data.json", function(data) {
      // Now data is the object {x:1, y:2} }
      );
      ```

- jQuery.get() and jQuery.post():  They both take the same three arguments as jQuery.getJSON(): a required URL, an optional data string or object, and a techni- cally optional but almost always used callback function.
- jQuery.ajax(): Common OPtions
     - type: http method, "GET", "POST"
     - url
     - data: data to e append to the url(get method) or send in the body of post request
     - dataType: Specifies the type of data expected in the response and how that data should be processed by jQuery. Legal values are “text”, “html”, “script”, “json”, “jsonp”, and “xml”.<Paste>
     - contentType: This specifies the HTTP Content-Type header for the re- quest. The default is “application/x-www-form- urlencoded”, which is the normal value used by HTML forms and most server-side scripts.
     - timeout: milliseconds
     - cache: For GET requests, if this option is set to false, jQuery will add a _= parameter to the URL or replace an existing pa- rameter with that name.
     - ifModified: When this option is set to true, jQuery records the values of the Last-Modified and If-None-Match response headers for each URL it requests, and then sets those headers in any subsequent requests for the same URL.
     - global: This option specifies whether jQuery should trigger events that describe the progress of the Ajax request.

- callbacks:
    - context: This option specifies the object to be used as the context —the this value—for invocations of the various callback functions.
    - beforeSend: This option specifies a callback function that will be in- voked before the Ajax request is sent to the server.
    - sucess: callback(data, statusCode, XMLHttpRequest) This option specifies the callback function to be invoked when an Ajax request completes successfully.
    - error: callback(XMLHttpRequest, statuscode(str))
    - complete: callback(XMLHttpRequest, statusCode) jQuery invokes the complete call- back after invoking either success or error.


Uncommon Options and hooks:
    - async
    - dataFilter
    - jsonp
    - jsonpCallback
    - processData
    - scriptCharset
    - traditional
    - username, password
    - xhr

- Ajax Events:  beforeSend, success, error, and complete.  remember that you can prevent jQuery from triggering any Ajax-related events by setting the global option to false.


# 7. Utility Functions
- jQuery.browser : object
- jQuery.contains(document_elements1, document_elements2)
- jQuery.each(): the jQuery.each() utility function iterates through the elements of an array or the properties of an object.
- jQuery.extend():  It copies the properties of the second and subsequent objects into the first object, overwriting any properties with the same name in the first argument. 优点类似于python 里的字典的 update
- jQuery.globalEval(): This function executes a string of JavaScript code in the global context, as if it were the contents of a <script> tag.
- jQuery.grep(): This function is like the ES5 filter() method of the Array object.
- jQuery.inArray(): is like the ES5 indexOf()
- jQuery.isArray(): Returns true if the argument is a native Array object.
- jQuery.isEmptyObject
- jQuery.isFunction()
- jQuery.isPlainObject()
- jQuery.makeArray()
- jQuery.map()
- jQuery.merge() : var clone = jQuery.merge([], original);
- jQuery.parseJSON()
- jQuery.proxy()
- jQuery.support
- jQuery.trim()


- # 8. Selectors and Selection Methods
- Simple Selectors: css 选择符
- Selector Combinaitons: processed left-to-right.
  - A B   descendants
  - A > B   direct-children
  - A + B  follow (ignoring text nodes and comments) elements
  - A ~ B  sibling elements that come after elements that match A.
- Selector Groups: "h1, h2, h3" A selector group matches all elements that match any of the selector combinations in the group.
- Selection Methods:
  - single(): jQuery object that contains only the single selected element at the specified index. 
  - last()
  - eq()
  - slice()
  - filter(): $("div").filter(function(i) { // Like $("div:even") return i % 2 == 0 })
  - not(): $("div").not("#header, #footer"); // All <div> tags except two special ones

- Using a Selection As Context
  - find()
  - children():  method returns the immediate child elements of each selected element, filtering them with an optional selector:
    - $("#header, #footer").children("span")  <==>  $("#header>span, #footer>span")
  - contents() method is similar to children() but it returns all child nodes, including text nodes, of each element.
  - next()
  - prev(): The next() and prev() methods return the next and previous sibling of each selected element that has one. 
  - nextAll() and prevAll() return all siblings following and pre- ceding (if there are any) each selected element. 
  - siblings() method returns all siblings of each selected element
  - parent()
  - closest()
- Reverting to a Previous Selection: To facilitate method chaining, most jQuery object methods return the object on which they are called. However, the meth- ods we’ve covered in this section all return new jQuery objects.
  - end(): To facilitate method chaining, most jQuery object methods return the object on which they are called. However, the meth- ods we’ve covered in this section all return new jQuery objects.
  - push Stack()
  -  and Self() returns a new jQuery object that includes all of the el- ements of the current selection plus all of the elements (minus duplicates) of the previous selection

# 9 Extending jQuery with Plugins
The trick is to know that jQuery.fn is the prototype object for all jQuery objects.


# 10 The jQuery UI Library
 http://jqueryui.com

# 11 jQuery Quick Reference
