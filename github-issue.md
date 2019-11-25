{{ $issueData := . }}
## Welcome Aboard

{{if or $issueData.AuditReportUrl $issueData.ReleaseReportUrl $issueData.PackageReportUrl}}

This source code repository has been configured as an application in [Sonatype Nexus IQ](https://guides.sonatype.com/iqserver/technical-guides/iq-server-for-devs/?utm_source=github&utm_medium=github-issue&utm_campaign=ce-iq-promo)

{{if $issueData.AuditReportUrl}}

The GitHub reported dependencies have been audited against your organization's policy. To
view the results of this audit, navigate to:

[Application Report - GitHub Dependency Audit]({{$issueData.AuditReportUrl}})

{{end}}
{{if $issueData.ReleaseReportUrl}}

Your latest release has been scanned and evaluated using the Nexus IQ CLI. To view
the results of this comprehensive evaluation, navigate to:

[Application Report - Release Evaluation]({{$issueData.ReleaseReportUrl}})

{{end}}
{{if $issueData.ReleaseReportUrl}}

Your latest package has been scanned and evaluated using the Nexus IQ CLI. To view
the results of this comprehensive evaluation, navigate to:

[Application Report - Package Evaluation]({{$issueData.PackageReportUrl}})

{{end}}

If you need help accessing a report, please contact:

[{{$issueData.Contact}}](mailto:{{$issueData.Contact}}?subject=Nexus%20IQ%20Access%20For%20{{$issueData.Repository}})

Now that you have an audit of the data provided by GitHub, configure Nexus IQ in your CI
to get continuous evaluations of your application's components.

{{else}}

This source code repository has been configured as an application in [Sonatype Nexus IQ](https://guides.sonatype.com/iqserver/technical-guides/iq-server-for-devs/?utm_source=github&utm_medium=github-issue&utm_campaign=ce-iq-promo). In order to
get continuous evaluations of your application's components configure Nexus IQ in your CI. To get
access to Nexus IQ, please contact:

[{{$issueData.Contact}}](mailto:{{$issueData.Contact}}?subject=Nexus%20IQ%20Access%20For%20{{$issueData.Repository}})

{{end}}

Want to get a deep dive into how Nexus IQ can streamline component usage for developers? Check out [IQ for Developers 100](https://learn.sonatype.com/courses/course-v1:Sonatype+IQ-DEV-100+01/about?utm_source=github&utm_medium=github-issue&utm_campaign=ce-iq-promo).

### Jenkins Configuration

<details>
    <summary>Click to see how to setup Nexus IQ for Jenkins</summary>

1. Install the [Nexus Platform Plugin](https://plugins.jenkins.io/nexus-jenkins-plugin)
1. Configure your Nexus IQ Server to {{$issueData.IqServerUrl}}
1. Configure authorization to Nexus IQ Server
1.  Add the policy evaluation to your pipeline after your application is built. Each `iqScanPattern` is an [Ant styled](https://ant.apache.org/manual/dirtasks.html) selector for files or archives to evaluate.       
```
nexusPolicyEvaluation iqApplication: '{{$issueData.Repository}}', iqScanPatterns: [[scanPattern: '**/*.js'], [scanPattern: '**/*.zip'], [scanPattern: '**/*.tar.gz'], [scanPattern: '**/*.jar'], [scanPattern: '**/*.war']], iqStage: 'build'
```

Using Freestyle builds or looking for ways to customize your evaluation? Check out [Sonatype Help](https://help.sonatype.com/integrations/nexus-and-continuous-integration/nexus-platform-plugin-for-jenkins).
</details>

### Bamboo Configuration

<details>
    <summary>Click to see how to setup Nexus IQ for Bamboo</summary>
    
1. Install [Nexus IQ for Bamboo](https://help.sonatype.com/iqserver/product-information/download-and-compatibility#DownloadandCompatibility-Bamboo)
1. Configure your Nexus IQ Server to {{$issueData.IqServerUrl}}
1. Configure authorization to Nexus IQ Server
1. Add an IQ Policy Evaluation task after your application is built with:
    1. `Application`: {{$issueData.Repository}}
    1. `Stage`: Build
    1. `Scan Targets`: [Ant style](https://ant.apache.org/manual/dirtasks.html) comma seperated list of files to evaluate. (e.g. **/*.js, **/*.zip, **/*.tar.gz, **/*.jar, **/*.war)
    
Looking for more details or ways to customize your evaluation? Check out [Sonatype Help](https://help.sonatype.com/integrations/nexus-and-continuous-integration/nexus-iq-for-bamboo).
</details>

### GitHub Actions Configuration

<details>
    <summary>Click to see how to setup Nexus IQ for GitHub Actions</summary>
    
1. Add [Nexus IQ for GitHub Actions](https://github.com/marketplace/actions/nexus-iq-for-github-actions) to your workflow
1. Configure workflow step with:
    1. `serverUrl`: {{$issueData.IqServerUrl}}
    1. `username`: Your IQ username
    1. `password`: Your password
    1. `application`: {{$issueData.Repository}}
    1. `stage`: Build
    1. `target`: The path to a specific application archive file, a directory containing such archives or the ID of a Docker image. For archives, a number of formats are supported, including jar, war, ear, tar, tar.gz, zip and many others.
</details>

### CLI Configuration

<details>
    <summary>Click to see how to setup Nexus IQ for CLI</summary>

The Nexus IQ CLI can be used in many build systems that allow for running CLIs via script.

1. Download the [Nexus IQ CLI](https://help.sonatype.com/iqserver/product-information/download-and-compatibility#DownloadandCompatibility-IQServer&CLI)
1. Evaluate an Archive or Directory with:
```
java -jar [nexus-iq-cli jar] -i {{$issueData.Repository}} -s {{$issueData.IqServerUrl}} -a [username:password] [target]
```

Looking for more details or ways to customize your evaluation?  Check out [Sonatype Help](https://help.sonatype.com/integrations/nexus-iq-cli).
</details>