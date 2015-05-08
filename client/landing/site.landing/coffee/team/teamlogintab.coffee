JView           = require './../core/jview'
CustomLinkView  = require './../core/customlinkview'
MainHeaderView  = require './../core/mainheaderview'
LoginInlineForm = require './../login/loginform'

module.exports = class TeamLoginTab extends KDTabPaneView

  JView.mixin @prototype

  constructor:(options = {}, data)->

    options.name = 'login'

    super options, data

    { mainController } = KD.singletons

    @header = new MainHeaderView
      cssClass : 'team'
      navItems : [
        { title : 'Blog',        href : 'http://blog.koding.com',   name : 'blog' }
        { title : 'Features',    href : '/Features',                name : 'features' }
      ]

    @logo = new KDCustomHTMLView tagName : 'figure'

    # set this once uploader ready - SY
    # @logo.setCss 'background-image', KD.config.group.backgroundImage

    @loginForm = new LoginInlineForm
      cssClass : 'login-form clearfix'
      testPath : 'login-form'
      callback : (formData) =>
        mainController.on 'LoginFailed', => @loginForm.button.hideLoader()
        mainController.login formData

    @loginForm.button.unsetClass 'solid medium green'
    @loginForm.button.setClass 'TeamsModal-button TeamsModal-button--green'

    if location.search isnt '' and location.search.search('username=') > 0
      username = location.search.split('username=').last.replace(/\&.+/, '')
      @loginForm.username.input.setValue username
      @loginForm.username.inputReceivedKeyup()



    @invitationLink = new CustomLinkView
      cssClass    : 'invitation-link'
      title       : 'Ask for an invite'
      testPath    : 'landing-recover-password'
      href        : '/Recover'

  pistachio: ->

    """
    {{> @header }}
    <div class="TeamsModal">
      {{> @logo}}
      <h4>Sign in to #{KD.config.group.title}</h4>
      {{> @loginForm}}
    </div>
    <section>
      <p>
      To be able to login to #{KD.config.groupName}.koding.com, you need to be invited by team administrators.
      </p>
      <p>
      Trying to create a team? <a href="/Teams">Sign up on the home page</a> to get started.
      </p>
    </section>
    <footer>
      <a href="/Legal" target="_blank">Acceptable user policy</a><a href="/Legal/Copyright" target="_blank">Copyright/DMCA guidelines</a><a href="/Legal/Terms" target="_blank">Terms of service</a><a href="/Legal/Privacy" target="_blank">Privacy policy</a>
    </footer>
    """