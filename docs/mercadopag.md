# Criar aplicação

As **aplicações** são entidades registradas dentro do Mercado Pago que atuam como um identificador único para gerenciar a autenticação e a autorização de suas integrações. Ou seja, são o vínculo entre seu desenvolvimento e o Mercado Pago e constituem como a primeira etapa para realizar a integração.

Com elas, é possível acessar as :toolTipComponent[credenciais]{link="/developers/pt/docs/credentials" linkText="Credenciais" content="Chaves de acesso únicas que usamos para identificar uma integração na sua conta, estando vinculadas à sua aplicação. Para mais informações, acesse o link abaixo."} necessárias para interagir com nossas APIs ou serviços específicos, além de gerenciar e organizar sua integração. Por isso, será necessário criar uma aplicação para cada solução do Mercado Pago integrada.

Para criar uma **aplicação**, siga os passos abaixo.

1. No canto superior direito de Mercado Pago Developers, clique em **Entrar** e faça login em sua conta do Mercado Pago.
2. Com a sessão iniciada, no canto superior direito do Mercado Pago Developers, clique em **Criar aplicação** se a sua conta ainda não tiver nenhuma aplicação criada, ou acesse "Suas integrações" e selecione **Ver todas**. Nela, clique em **Criar aplicação**.
3. Uma vez dentro de **Suas integrações**, clique no botão **Criar aplicação**.

> NOTE
>
> Para proteger sua conta e garantir a conformidade das operações, durante a criação de uma aplicação será necessário que realize uma verificação de identidade, caso ainda não tenha sido feita, ou uma reautenticação, se já tiver concluído previamente o processo de verificação.

![create-application-1](/images/snippets/create-application-1-pt-v1.png)

4. Insira um **nome** para identificar sua aplicação. O limite é de até 50 caracteres alfanuméricos.
5. Selecione **Pagamentos online** como o tipo de pagamento que você deseja integrar, pois essa é a opção correspondente a lojas virtuais. Clique em **Continuar**.
6. Indique que você está integrando para uma loja desenvolvida por conta própria. Opcionalmente, você poderá preencher a URL da sua loja. Clique em **Continuar**.
7. Selecione a opção **Checkouts** e, em seguida, escolha **Checkout Transparente** como a solução que você vai integrar.
8. Selecione **API de Orders** como o tipo de API com o qual você vai integrar e depois selecione **Continuar**.
9. Confirme as opções selecionadas. Caso você precise modificar alguma seleção, clique no botão **Editar**. Aceite a [Declaração de Privacidade](https://www.mercadopago.com.br/privacidade) e os [Termos e Condições](/developers/pt/docs/resources/legal/terms-and-conditions) e clique em **Confirmar**.

![Resumo de aplicação](/images/snippets/create-application/PT-new-app-API-de-orders-v1.png)

Em [Suas integrações](/developers/panel/app), é possível consultar a lista de todas as suas aplicações criadas e acessar os [Dados da integração](/developers/pt/docs/checkout-api-orders/resources/application-details) de cada uma delas.

> NOTE
>
> Caso necessário, é possível editar ou excluir uma aplicação. Neste último caso, tenha em mente que sua loja deixará de receber pagamentos por meio da integração com o Mercado Pago associada a essa aplicação. Para mais informações, acesse os [Dados da integração](/developers/pt/docs/checkout-api-orders/resources/application-details).

## Acessar credenciais de teste

Após criar a sua aplicação, também serão automaticamente criadas as :toolTipComponent[credenciais de teste]{link="/developers/pt/docs/checkout-api-orders/resources/credentials" linkText="Credenciais" content="Chaves de acesso únicas que usamos para identificar uma integração na sua conta, estando vinculadas à sua aplicação Para mais informações, acesse o link abaixo."}. Utilize as **credenciais de teste** para realizar todas as configurações e validações necessárias em um ambiente seguro de testes.

Ao acessar as credenciais de teste, serão exibidos os seguintes pares de credenciais: :toolTipComponent[Public Key]{content="Chave pública que é utilizada no _frontend_ para acessar informações e criptografar dados. Você pode acessá-la através de *Suas integrações > Dados da integração > Testes > Credenciais de teste*."} e o :toolTipComponent[Access Token de teste]{content="Chave privada da aplicação criada no Mercado Pago e que é utilizada no _backend_. Você pode acessá-la através de *Suas integrações > Dados da integração > Testes > Credenciais de teste*."}. O _Access Token_ de teste começa com o prefixo `APP_USR`, assim como o seu _Access Token_ de produção.

![credenciais de teste](/images/snippets/credentials/application-data-test-credentials-pt-v1.png)

> NOTE
>
> Se você estiver utilizando uma aplicação já existente, será necessário ativar as credenciais de teste. Para mais informações, acesse a documentação de [Credenciais](/developers/pt/docs/checkout-api-orders/resources/credentials).

# Configurar ambiente de desenvolvimento 

Para começar a integrar as soluções de pagamento do Mercado Pago, é necessário preparar seu ambiente de desenvolvimento com uma série de configurações básicas que permitirão acessar as funcionalidades do Mercado Pago a partir do _frontend_ de forma segura.

O ambiente de desenvolvimento pode ser configurado para integrações em **websites** ou **aplicativos mobile**. Selecione o tipo de integração abaixo e siga as instruções específicas para cada caso.

:::::TabsComponent
::::TabComponent{title="Websites"}
> CLIENT_SIDE
>
> h2
>
> Incluir a biblioteca MercadoPago.js

Utilize nossas bibliotecas oficiais para acessar as funcionalidades do Mercado Pago a partir do seu _frontend_ e capturar os dados dos pagamentos realizados de forma segura.

[[[
```html
<body>
  <script src="https://sdk.mercadopago.com/js/v2"></script>
</body>
```
```node
npm install @mercadopago/sdk-js

```
]]]

> CLIENT_SIDE
>
> h2
>
> Inicializar biblioteca do Mercado Pago

Para inicializar a biblioteca do Mercado Pago, é necessário utilizar suas :toolTipComponent[credenciais]{link="/developers/pt/docs/checkout-api-orders/resources/credentials" linkText="Credenciais" content="Chaves de acesso únicas que usamos para identificar uma integração na sua conta, estando vinculadas à sua aplicação. Para mais informações, acesse o link abaixo."}, chaves únicas que identificam uma integração na sua conta. Estão diretamente vinculadas à :toolTipComponent[aplicação]{link="/developers/pt/docs/checkout-api-orders/resources/application-details" linkText="Dados da integração" content="Entidade registrada no Mercado Pago que atua como um identificador para gerenciar suas integrações. Para mais informações, acesse o link abaixo."} criada para essa integração, permitindo que você desenvolva seu projeto contando com as melhores medidas de segurança do Mercado Pago.

Nesta etapa, você deverá usar sua :toolTipComponent[Public Key de teste]{content="Chave pública que é utilizada no _frontend_ para acessar informações e criptografar dados. Você pode acessá-la através de *Suas integrações > Dados da integração*, indo até a seção *Credenciais*, localizada à direita da tela, e clicando em *Teste*. Alternativamente, você também poderá acessá-la a partir de *Suas integrações > Dados da aplicação > Testes > Credenciais de teste*."}.

> NOTE
>
> Se você estiver desenvolvendo para outra pessoa, poderá acessar as credenciais das aplicações que você não gerencia. Consulte [Compartilhar credenciais](/developers/pt/docs/checkout-api-orders/resources/credentials#bookmark_compartilhar_credenciais) para mais informações.

Uma vez localizada a :toolTipComponent[Public Key de teste]{content="Chave pública que é utilizada no _frontend_ para acessar informações e criptografar dados, seja na etapa de desenvolvimento ou na de testes. Você pode acessá-la através de *Suas integrações > Dados da integração*, indo até a seção *Credenciais* l,ocalizada à direita da tela. e clicando em *Teste*. Alternativamente, você também poderá acessá-la a partir de *Suas integrações > Dados da integração > Testes > Credenciais de teste*."}, copie-a e a inclua no _frontend_ para poder acessar os dados necessários para interagir com nossos serviços, assim como criptografar aqueles dados sensíveis envolvidos nos pagamentos que você irá receber.

[[[
```html
<script>
  const mp = new MercadoPago("YOUR_PUBLIC_KEY");
</script>
```
```javascript
import { loadMercadoPago } from "@mercadopago/sdk-js";

await loadMercadoPago();
const mp = new window.MercadoPago("YOUR_PUBLIC_KEY");
```
]]]

Com essas configurações, seu ambiente de desenvolvimento já está pronto para continuar com as configurações específicas da sua integração.

::::
::::TabComponent{title="Aplicativos mobile"}

Utilize o SDK nativo do Mercado Pago e simplifique o processo de pagamento via cartão em aplicativos Android e iOS de forma segura. Selecione a tecnologia utilizada e siga as etapas para configurar o ambiente de desenvolvimento e iniciar o fluxo de pagamento.

:::AccordionComponent{title="iOS"}

Utilize o SDK nativo do Mercado Pago para integrar meios de pagamento em aplicativos iOS. Veja a seguir como realizar a instalação e a inicialização do SDK.

### Instalar SDK 

Confira abaixo o passo a passo para instalar o SDK no seu projeto Swift.

1. No Swift Package Manager, clique em **Arquivo > Adicionar pacotes**.
2. Cole a URL do repositório: `https://github.com/mercadopago/sdk-ios`.
3. Selecione a versão desejada do SDK.
4. Clique em **Adicionar pacote** para concluir a instalação.

### Adicionar dependências

Importe as dependências do SDK no seu projeto executando o seguinte código:

```
import CoreMethods
```

### Inicializar SDK

Após instalar o SDK e adicionar as dependências ao seu projeto, inicialize o SDK no início do ciclo de vida do aplicativo. Isso garante que todas as configurações essenciais estejam definidas antes de qualquer operação de pagamento.

> WARNING
>
> O SDK deve ser inicializado uma única vez no momento da abertura do aplicativo. Para garantir o funcionamento correto, realize uma chamada a `initialize()` antes de utilizar qualquer outra funcionalidade do SDK.

Para inicializar a biblioteca do Mercado Pago, é necessário utilizar suas :toolTipComponent[credenciais]{link="/developers/pt/docs/checkout-api-orders/resources/credentials" linkText="Credenciais" content="Chaves de acesso únicas que usamos para identificar uma integração na sua conta, estando vinculadas à sua aplicação. Para mais informações, acesse o link abaixo."}, chaves únicas que identificam sua integração e estão vinculadas à :toolTipComponent[aplicação]{link="/developers/pt/docs/checkout-api-orders/resources/application-details" linkText="Dados da integração" content="Entidade registrada no Mercado Pago que atua como um identificador para gerenciar suas integrações. Para mais informações, acesse o link abaixo."} criada para essa integração, permitindo que você desenvolva seu projeto contando com as melhores medidas de segurança do Mercado Pago.

Nesta etapa, você deverá usar sua :toolTipComponent[Public Key de teste]{content="Chave pública que é utilizada no _frontend_ para acessar informações. Você pode acessá-la através de *Suas integrações > Dados da integração*, indo até a seção *Credenciais*, localizada à direita da tela, e clicando em *Teste*. Alternativamente, você também poderá acessá-la a partir de *Suas integrações > Dados da aplicação > Testes > Credenciais de teste*."}.

> NOTE
>
> Se você estiver desenvolvendo para outra pessoa, poderá acessar as credenciais das aplicações que você não gerencia. Consulte [Compartilhar credenciais](/developers/pt/docs/checkout-api-orders/resources/credentials#bookmark_compartilhar_credenciais) para mais informações.

Copie a :toolTipComponent[Public Key de teste]{content="Chave pública que é utilizada no _frontend_ para acessar informações. Você pode acessá-la através de *Suas integrações* na seção *Dados da integração*, indo até a seção *Credenciais* localizada à direita da tela e clicando em *Teste*. Alternativamente, você pode acessar por *Suas integrações > Dados da aplicação > Testes > Credenciais de teste*."} e a inclua no código abaixo. O processo de inicialização varia conforme a tecnologia utilizada, seja UIKit ou SwiftUI.

[[[
```UIKit
import UIKit
import CoreMethods

@main
class AppDelegate: UIResponder, UIApplicationDelegate {
  func application(_ application: UIApplication, 
  didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
  ) -> Bool {
  let configuration = MercadoPagoSDK.Configuration(
  publicKey: "YOUR-PUBLIC-KEY",
  country: // Insira o país da sua chave pública
  )
  MercadoPagoSDK.shared.initialize(configuration)
  
  return true
  }
}
```
```SwiftUI

import SwiftUI
import CoreMethods

@main
struct YourApp: App {
  init() {
  let configuration = MercadoPagoSDK.Configuration(
  publicKey: "<YOUR-PUBLIC-KEY>",
  country: "<Insira o país da sua chave pública>",
  locale: "pt-BR"
  )
  MercadoPagoSDK.shared.initialize(configuration)
  }
  
  var body: some Scene {
  WindowGroup {
  ContentView()
  }
  }
}
```
]]]

Os parâmetros de inicialização estão listados na tabela abaixo.

| Parâmetro | Tipo | Descrição | Obrigatoriedade |
| ------------ | ------- | ------------------------------------------------------------------------------------------- | -------------- |
| `public_key` | String | Chave pública que é utilizada no _frontend_ para acessar informações. Você pode acessá-la através de **Suas integrações > Dados da integração > Testes > Credenciais de teste**. | Obrigatório |
| `locale` | String | Identificador do _locale_ (idioma e país). Por padrão, utiliza-se o _locale_ do sistema. | Opcional |
| `country` | [Country](https://mercadopago.github.io/sdk-ios/0.1.0/documentation/coremethods/mercadopagosdk/country) | `Enum` que identifica o país em que os _Core Methods_ serão processados. Utilize o código do país correspondente à sua Public Key. Consulte a [documentação](https://mercadopago.github.io/sdk-ios/0.1.0/documentation/coremethods/mercadopagosdk/country/) para verificar o código referente ao seu país. | Obrigatório |

Com essas configurações, seu ambiente de desenvolvimento já está pronto para continuar com as configurações específicas da sua integração.
:::

:::AccordionComponent{title="Android"}

Utilize o SDK nativo do Mercado Pago para integrar meios de pagamento em aplicativos Android. Veja a seguir como realizar a configuração do repositório e a inicialização do SDK.

### Configurar repositório

Adicione o repositório do Mercado Pago ao arquivo `settings.build.gradle` do seu projeto em Kotlin, conforme o exemplo abaixo:

```kotlin
pluginManagement {
  repositories {
  // Outras dependências...
  maven { url = uri("https://artifacts.mercadolibre.com/repository/android-releases") }
  }
}
dependencyResolutionManagement {
  repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
  repositories {
  // Outras dependências...
  maven {
  url = uri("https://artifacts.mercadolibre.com/repository/android-releases")
  }
  }
}
```

### Adicionar dependências

Inclua as dependências do SDK no arquivo `build.gradle` (ou `build.gradle.kts`) do módulo do seu aplicativo:

```kotlin
// Especifique o SDK BOM com uma definição de versão
implementation(platform("com.mercadopago.android.sdk:sdk-android-bom:<ultima versão do bom>"))
// Especifique as dependências da biblioteca SDK sem uma definição de versão
implementation("com.mercadopago.android.sdk:core-methods")
```

### Inicializar SDK

Após configurar o repositório e adicionar as dependências ao seu projeto, inicialize o SDK no início do ciclo de vida do aplicativo. Isso garante que todas as configurações essenciais estejam definidas antes de qualquer operação de pagamento.

> WARNING
>
> O SDK deve ser inicializado uma única vez no momento da abertura do aplicativo. Para garantir o funcionamento correto, realize uma chamada a `initialize()` antes de utilizar qualquer outra funcionalidade do SDK.

Para inicializar a biblioteca do Mercado Pago, é necessário utilizar suas :toolTipComponent[credenciais]{link="/developers/pt/docs/checkout-api-orders/resources/credentials" linkText="Credenciais" content="Chaves de acesso únicas que usamos para identificar uma integração na sua conta, estando vinculadas à sua aplicação. Para mais informações, acesse o link abaixo."}, que estão vinculadas à :toolTipComponent[aplicação]{link="/developers/pt/docs/checkout-api-orders/resources/application-details" linkText="Dados da integração" content="Entidade registrada no Mercado Pago que atua como um identificador para gerenciar suas integrações. Para mais informações, acesse o link abaixo."} que você criou.

Nesta etapa, você deverá usar sua :toolTipComponent[Public Key de teste]{content="Chave pública que é utilizada no _frontend_ para acessar informações. Você pode acessá-la através de *Suas integrações > Dados da integração*, indo até a seção *Credenciais*, localizada à direita da tela, e clicando em *Teste*. Alternativamente, você também poderá acessá-la a partir de *Suas integrações > Dados da aplicação > Testes > Credenciais de teste*."}.

> NOTE
>
> Se você estiver desenvolvendo para outra pessoa, poderá acessar as credenciais das aplicações que você não gerencia. Consulte [Compartilhar credenciais](/developers/pt/docs/checkout-api-orders/resources/credentials#bookmark_compartilhar_credenciais) para mais informações.

Copie a :toolTipComponent[Public Key de teste]{content="Chave pública que é utilizada no _frontend_ para acessar informações. Você pode acessá-la através de *Suas integrações > Dados da integração*, indo até a seção *Credenciais*, localizada à direita da tela, e clicando em *Teste*. Alternativamente, você também poderá acessá-la a partir de *Suas integrações > Dados da aplicação > Testes > Credenciais de teste*."} e a inclua no código abaixo. Após isso, inicialize o SDK na classe `Application`, conforme o exemplo:

```kotlin
import android.app.Application
import com.mercadopago.sdk.android.initializer.MercadoPagoSDK

class MainApplication : Application() {
  override fun onCreate() {
  super.onCreate()
  MercadoPagoSDK.initialize(
  context = this,
  publicKey = "<YOUR-PUBLIC-KEY>",
  countryCode = //CountryCode of this public key
  )
  }
}
```

Os parâmetros de inicialização estão listados na tabela abaixo.

| Parâmetro | Tipo | Descrição | Obrigatoriedade |
| ------------ | ------- | ---------------------------------------------------------------------------------------------------------------- | -------------- |
| `context` | [Context](https://developer.android.com/reference/android/content/Context) | Contexto da sua aplicação. | Obrigatório |
| `publicKey` | String | Chave pública que é utilizada no _frontend_ para acessar informações. Você pode acessá-la através de **Suas integrações > Dados da integração > Testes > Credenciais de teste**. | Obrigatório |
| `countryCode`| [CountryCode](https://mercadopago.github.io/sdk-android/sdk-android/com.mercadopago.sdk.android.domain.model/-country-code/index.html?query=enum%20CountryCode%20:%20Enum%3CCountryCode%3E) | `Enum` que identifica o país em que os _Core Methods_ serão processados. Utilize o código do país correspondente à sua Public Key. Consulte a [documentação](https://mercadopago.github.io/sdk-ios/0.1.0/documentation/coremethods/mercadopagosdk/country/) para verificar o código referente ao seu país. | Obrigatório |

Com essas configurações, seu ambiente de desenvolvimento já está pronto para continuar com as configurações específicas da sua integração.
:::

::::
:::::