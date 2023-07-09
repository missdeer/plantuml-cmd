require 'open3'
require 'fileutils'

module Jekyll

  class PlantUMLGanttBlock < Liquid::Block
    attr_reader :config
    
    def render(context)
      site = context.registers[:site]
      self.config = site.config['plantuml']
      
      tmproot = File.expand_path(tmp_folder)
      folder = "/images/"
      create_tmp_folder(tmproot, folder)

      code = super
      filename = Digest::MD5.hexdigest(code) + ".svg"
      filepath = tmproot + folder + filename
      if !File.exist?(filepath)
        plantuml_cmd = File.expand_path(plantuml_cmd_path)
        if config['remote'].eql? "enabled"
          cmd = plantuml_cmd + " -r -t gantt -o " + filepath
        else
          cmd = plantuml_cmd + " -t gantt -o " + filepath
        end
        result, status = Open3.capture2e(cmd, :stdin_data=>code)
        Jekyll.logger.debug(filepath + " -->\t" + status.inspect() + "\t" + result)
      end
      
      text = File.read(filepath)
      startPos = text.index('<svg')
      source = text[startPos..-1]
      source = source.gsub('font-family="sans-serif"', 'font-family=""')
    end

    private

    def config=(cfg)
      @config = cfg || Jekyll.logger.abort_with("Missing 'plantuml' configurations.")
    end
        
    def plantuml_cmd_path
      config['plantuml_cmd'] || Jekyll.logger.abort_with("Missing configuration 'plantuml.plantuml_cmd'.")
    end
    
    def tmp_folder
      config['tmp_folder'] || Jekyll.logger.abort_with("Missing configuration 'plantuml.tmp_folder'.")
    end
        
    def create_tmp_folder(tmproot, folder)
      folderpath = tmproot + folder
      if !File.exist?(folderpath)
        FileUtils::mkdir_p folderpath
        Jekyll.logger.info("Create PlantUML image folder: " + folderpath)
      end
    end
    
  end # PlantUMLGanttBlock
end

Liquid::Template.register_tag('gantt', Jekyll::PlantUMLGanttBlock)
