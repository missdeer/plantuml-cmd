require 'open3'
require 'fileutils'

module Jekyll

  class PlantUMLBlock < Liquid::Block
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
        cmd = plantuml_cmd + " -o " + filepath
        result, status = Open3.capture2e(cmd, :stdin_data=>code)
        Jekyll.logger.debug(filepath + " -->\t" + status.inspect() + "\t" + result)
      end
      
      text = File.read(filepath)
      startPos = text.index('<svg')
      source = text[startPos..-1]
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
    
  end # PlantUMLBlock
end

Liquid::Template.register_tag('plantuml', Jekyll::PlantUMLBlock)
