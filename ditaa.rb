require 'open3'
require 'fileutils'

module Jekyll

  class PlantUMLDitaaBlock < Liquid::Block
    attr_reader :config
    
    def render(context)
      site = context.registers[:site]
      self.config = site.config['plantuml']
      
      tmproot = File.expand_path(tmp_folder)
      folder = "/images/"
      create_tmp_folder(tmproot, folder)

      code = super
      filename = Digest::MD5.hexdigest(code) + ".png"
      filepath = tmproot + folder + filename
      if !File.exist?(filepath)
        plantuml_cmd = File.expand_path(plantuml_cmd_path)
        cmd = plantuml_cmd + " -t ditaa -f png -o " + filepath
        result, status = Open3.capture2e(cmd, :stdin_data=>code)
        Jekyll.logger.debug(filepath + " -->\t" + status.inspect() + "\t" + result)
      end

      site.static_files << Jekyll::StaticFile.new(site, tmproot, folder, filename)
      
      source = "<img src='" + site.config['baseurl'] + folder + filename + "'>"
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
    
  end # PlantUMLDitaaBlock
end

Liquid::Template.register_tag('ditaa', Jekyll::PlantUMLDitaaBlock)
