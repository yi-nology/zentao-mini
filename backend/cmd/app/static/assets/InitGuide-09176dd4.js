import{_ as g,K as c,L as h,r as b,b as r,d as l,m as t,M as m,t as i,n as d,l as a,e as p,f as I}from"./index-759f9a80.js";const _={name:"InitGuide",data(){return{selectedFile:null,loading:!1,error:"",success:"",testing:!1,testResult:"",debugInfo:""}},methods:{handleFileChange(s){this.selectedFile=s.target.files[0]},async submitForm(){if(!this.selectedFile){this.error="请选择加密配置文件";return}this.loading=!0,this.error="",this.success="",this.debugInfo="";try{const s=new FormData;s.append("configFile",this.selectedFile),this.debugInfo+=`请求URL: /api/init/upload
`,this.debugInfo+=`请求方法: POST
`,this.debugInfo+=`请求文件: ${this.selectedFile.name} (${this.selectedFile.size} bytes)
`,this.debugInfo+=`请求时间: ${new Date().toISOString()}

`;const e=await c(s);if(this.debugInfo+=`响应时间: ${new Date().toISOString()}

`,this.debugInfo+=`响应数据: ${JSON.stringify(e,null,2)}
`,e.code!==200)throw new Error(e.message||"初始化失败");localStorage.setItem("initialized","true"),this.success="初始化成功！系统已准备就绪。",setTimeout(()=>{this.$router.push("/")},3e3)}catch(s){this.error="初始化失败，请检查上传的文件并重试。",this.debugInfo+=`错误信息: ${s.message}
`,console.error("初始化错误:",s)}finally{this.loading=!1}},async testZentao(){this.testing=!0,this.testResult="",this.debugInfo="";try{this.debugInfo+=`请求URL: /api/users/current
`,this.debugInfo+=`请求方法: GET
`,this.debugInfo+=`请求时间: ${new Date().toISOString()}

`;const s=await h();if(this.debugInfo+=`响应时间: ${new Date().toISOString()}

`,this.debugInfo+=`响应数据: ${JSON.stringify(s,null,2)}
`,s.code!==200)throw new Error(s.message||"测试失败");this.testResult=JSON.stringify(s,null,2)}catch(s){this.testResult="测试失败: "+s.message,this.debugInfo+=`错误信息: ${s.message}
`,console.error("测试禅道连接错误:",s)}finally{this.testing=!1}}}},y={class:"init-guide"},F={class:"init-guide-container"},S={class:"form-group"},v={key:0,class:"file-info"},w={class:"form-actions"},k=["disabled"],C={key:0,class:"error-message"},O={key:1,class:"success-message"},R=["disabled"],D={key:2,class:"test-result"},N={key:3,class:"debug-info"};function x(s,e,E,G,n,o){const f=b("router-link");return r(),l("div",y,[t("div",F,[e[8]||(e[8]=t("h1",null,"系统初始化",-1)),e[9]||(e[9]=t("p",{class:"subtitle"},"请上传加密配置文件以完成系统初始化",-1)),t("form",{onSubmit:e[1]||(e[1]=m((...u)=>o.submitForm&&o.submitForm(...u),["prevent"])),class:"init-form"},[t("div",S,[e[3]||(e[3]=t("label",{for:"configFile"},"加密配置文件",-1)),t("input",{type:"file",id:"configFile",ref:"fileInput",onChange:e[0]||(e[0]=(...u)=>o.handleFileChange&&o.handleFileChange(...u)),accept:".json",required:""},null,544),e[4]||(e[4]=t("p",{class:"hint"},"请上传使用 generate-encryption.sh 脚本生成的 auth-config.json 文件",-1))]),n.selectedFile?(r(),l("div",v,[t("p",null,"已选择文件: "+i(n.selectedFile.name),1)])):d("",!0),t("div",w,[t("button",{type:"submit",class:"btn primary",disabled:n.loading||!n.selectedFile},i(n.loading?"初始化中...":"开始初始化"),9,k)])],32),n.error?(r(),l("div",C,i(n.error),1)):d("",!0),n.success?(r(),l("div",O,[a(i(n.success)+" ",1),p(f,{to:"/",class:"btn secondary"},{default:I(()=>[...e[5]||(e[5]=[a("进入系统",-1)])]),_:1}),t("button",{onClick:e[2]||(e[2]=(...u)=>o.testZentao&&o.testZentao(...u)),class:"btn secondary",disabled:n.testing},i(n.testing?"测试中...":"测试禅道连接"),9,R)])):d("",!0),n.testResult?(r(),l("div",D,[e[6]||(e[6]=t("h3",null,"测试结果",-1)),t("pre",null,i(n.testResult),1)])):d("",!0),n.debugInfo?(r(),l("div",N,[e[7]||(e[7]=t("h3",null,"调试信息",-1)),t("pre",null,i(n.debugInfo),1)])):d("",!0)])])}const V=g(_,[["render",x],["__scopeId","data-v-9cff2d68"]]);export{V as default};
