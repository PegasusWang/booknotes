Part1 The Mechanics of Change

# 1.Changing Software
Four primary ressons:
1. Adding a feature
2. Fixing a bug.
3. Improving the design
4. Optmmizing resoource usage

structure, functionality, resource usage

ask three questions:
1. What chagnes do we have to make?
2. How will we know that we've doen them correctly?
3. How will we know that we have't broken anything?


# 2. Working with Feedback
Changes in a system 2 ways:
Edit and Pray:
Cover and Modify: tests

The legacy code change algorith:
1. Identify change points.
2. Find test points.
3. Break dependencies.
4. Write tests.
5. Make changes and refactor.

# 3. Sensing and Sepatation
fake objects
mock objects

# 4.The Seam Model


# 5. Tools
Mock Objects
Unit-Tesiting Harnesses


Part2 The Mechanics of Change
# 6. Don't Have Much Time and I Have to Change It
Sprout method


# 7. It Takes Forever to make a change
减少依赖


# 8. How Do I Add a Feature?
TDD:
Programming by Difference

# 9. I Can't get this class into a test Harness
恼人的参数：如果构造函数的参数依赖其他对象， 可以通过构造fake对象作为参数传入，通过继承原对象并用空方法复写父类。
隐藏的依赖：不要在函数里隐藏依赖， 最好用函数参数的形式传入。
如果隐藏依赖过多，穿参数会导致大量参数，可以使用方法替换掉内部的实例变量。setter方法
全局依赖：对于单例模式等class，可以使用静态setter函数

# 11.I need to Make a Change, WHat methods should I test?
找到会受影响的代码


# 12.I need to make many changes in one Area. Do I have to break dependices for all the classes involved.
Interception Points: 由变更点决定。
把Interception point作为封装的边界

# 13. I need to make a change, but i don't know what tests to write.
Characterization test.

# 14. Dependencies on Libraies are killing me.

# 15. My application is all api callls
Responsibiliy-based extraction

# 16.I don't understand the code well enough to change it
Notes(Sketching): 通过画图和做笔记的形式理解代码，画出之间的连接和关系。
Lising Markup:  Separating Responsibilities; Understanding Method Structure; Extract Methods; Understand the Errects of a change.
Scratch Refacting:
Deleted Unused code

# 17. My Application Has no structure
Telling the Story of the system: 互相解说
Naked CRC: Class, Responsiblity and Collaborations. 1. Cards represent instances, not clases. 2. Overlap cards to show a collection of them
Conversation srutiny:

# 18. My Test Code is in the way
Class Naming Conventions: 对于测试类和Fake类通过后缀或者前缀来区别。
Test Location:

# 19. My Project is not Object Oriented. How do I make changes ?
通过预处理器宏添加测试
Adding New Behavior:

# 20. This class is Too big and I don't want it to get any bigger
不要写过大（方法太多）的类，会导致混淆、
Single-Responsibility Principle(SRP): It should have a single purpose in the system, and there should be only one reason to change it 。如何确定职责？功能相近的应该有类似后缀，根据功能划分。()
Seeing Responsibilities:
- 1. Group Methods(写出所有方法名和访问权限)
- 2. Look at Hidden Methods
- 3. Look for Deciions that can change.
- 4. Look for internal relationships（写出所有变量和方法使用的变量，画出调用关系，根据把调用关系分成的几个大区域来拆分class）.
- 5. Look for the primary responsibility.(尝试使用一句话描述类的职责) 接口隔离原则
- 6. When all else fails, do some scratch refactoring
- 7. Focus on the current work.


# 21. I'm Changing the same code all over the place
开闭原则

# 22. I need to change a Monster method and I can't wirte tests for it
Varieties of Monster: Bullete Methods(没有缩进); Snarled methods(单一大块代码段);
Tackling Monsters with automated refactoring support:
- To separate logic from awkward dependencies
- To introduce seams that make it easier to get tests in place for more refactoring

The manual refactoring challenge:


# 22. How Do I know That I'm Not Breaking Anything?
Hyperaware Editing
Single-Goal Editing: doing one thing at a time
Proserve Signatures
Lean on the Compiler
Pair Programming

# 24. We Feel Overwhelmed, It Isn't going to Get any better.
Finding what motivates you
如果因为代码质量团队士气低：找到最丑的代码类，然后让测试覆盖它。


# 25. Dependency-Breaking Techniques
