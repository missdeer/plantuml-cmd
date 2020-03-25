# plantuml-cmd
command line util to generate plantuml output via remote plantuml service

# Build && Install

```
go get github.com/missdeer/plantuml-cmd
```

# Usage

```
cat test.puml | ./plantuml-cmd
```

or

```
./plantuml-cmd -i test.puml
```

# Also Plugins for Jekyll

## Steps

1. Download all `.rb` files and put in `_plugins` directory
2. Add code lines shown below at the end of `_config.yml`:
```yaml
plantuml:
  remote: "enabled"
  plantuml_cmd: /usr/local/bin/plantuml-cmd   
  tmp_folder: _plantuml
```
Notice: make sure `plantuml-cmd` in the right place

## Sample Project

https://github.com/missdeer/blog/

## Sample Post

https://github.com/missdeer/blog/blob/gh-pages/_posts/2020-02-28-plantuml-sampels.md

## Live Sample Post

https://blog.minidump.info/2020/02/plantuml-sampels/
