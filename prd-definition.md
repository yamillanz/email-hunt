1. Problem Statement
Una o dos frases. Qué está roto o qué falta. Por qué importa ahora.

No es la solución. Es el problema.

“El módulo de autenticación guarda los tokens de sesión en localStorage. Legal ha identificado esto como un riesgo de compliance. Hay que migrar a httpOnly cookies antes del cierre del trimestre.”

Eso es un problem statement. Concreto, acotado, con consecuencias reales.

“Necesitamos mejorar la seguridad del sistema.”

Eso no es un problem statement. Es una aspiración.

2. Solution
Describe el destino. No el camino.

Qué tiene que ser verdad cuando el trabajo esté hecho, desde el punto de vista del usuario o del sistema.

“Los usuarios podrán autenticarse. Las sesiones persistirán entre recargas del navegador. Un logout invalida la sesión en el servidor y en el cliente. Los tokens nunca están accesibles desde JavaScript en el frontend.”

Nota lo que no dice: no dice JWT, no dice Redis, no dice qué librería usar. Dice qué tiene que ser verdad.

El agente decide el cómo. Tú decides el qué.

* * *
3. User Stories
Las acciones concretas que alguien realiza.

El formato clásico funciona:

Como [rol], cuando [acción], entonces [resultado esperado].

Tres o cuatro historias son suficientes para la mayoría de features. Si necesitas más de seis, probablemente estás especificando más de una feature.

* * *
4. Constraints
Las restricciones no negociables.

No preferencias. No sugerencias. Restricciones reales que, si se violan, el resultado es inválido independientemente de lo demás.

“Los tokens no pueden estar en localStorage ni sessionStorage bajo ninguna circunstancia.”

“La invalidación de sesión tiene que ocurrir en el servidor, no solo eliminando la cookie en el cliente.”

Si no tienes constraints reales, deja la sección vacía. No la inventes para rellenar.

5. Acceptance Criteria
Esto es lo más importante.

Cada criterio tiene que ser verificable. Tiene que poder pasar o fallar. Si no puede fallar, no es un criterio, es un deseo.

“Login con credenciales válidas → código 200 y cookie httpOnly en la respuesta.”

“Request con token expirado → código 401.”

“Logout → sesión invalidada en servidor, reutilizar el token devuelve 401.”

Cada uno de esos puede verificarse. Puedes hacer la prueba y decir: pasó o no pasó.

“El sistema funciona correctamente.”

No se puede verificar. No sirve.

* * *
6. Out of Scope
Lo que explícitamente no entra en este PRD.

Esta sección es la más ignorada. Y la que más problemas causa cuando falta.

Sin Out of Scope, el agente implementa todo lo relacionado con el tema. Si estás especificando autenticación, sin Out of Scope puede que el agente añada OAuth, recuperación de contraseña, 2FA, y bloqueo por intentos fallidos. Todo razonable. Nada de lo que pediste.

“OAuth y login social — fase 2.”

“2FA — issue separado.”

“Rate limiting en login — no entra en este PRD.”

Explícito. Sin ambigüedad.

* * *
7. Assumptions
La sección más crítica del PRD.

No es la más larga. Es la más importante.

El staff engineer de Augment Code lo dice sin rodeos: “La sección de assumptions es en la que más tiempo paso iterando. Es donde más impacto tiene el trabajo.”

¿Por qué?

Porque el agente toma decisiones basadas en lo que asume sobre tu sistema. Si esas assumptions son incorrectas, todo el downstream está mal. El código es técnicamente correcto pero para un sistema que no es el tuyo.

“El proyecto ya tiene Prisma configurado con tabla users.”

“El frontend es una React SPA separada del backend — no SSR.”

“No existen tests de integración actualmente.”

Si el agente asume que tienes SSR cuando tienes SPA, el enfoque para las cookies cambia por completo. Si asume que existen tests cuando no existen, puede que no los cree.

Las assumptions son donde la realidad de tu proyecto entra en la spec.

8. Implementation Notes
Opcional. Y muy diferente de las constraints.

Las constraints son no negociables. Las implementation notes son sugerencias.

“Considera si necesitas invalidación real en servidor — evalúa JWT vs sesiones en base de datos según el volumen esperado.”

El agente puede ignorarlas si tiene mejor razón. Eso es correcto. Si no las marcas claramente como sugerencias, el agente las tratará como obligaciones aunque sean malas ideas.

La diferencia entre “usa JWT” y “considera JWT” no es semántica. Es la diferencia entre prescribir y delegar.

# Ejemplo de PRD completo

# PRD: Migrar autenticación a httpOnly cookies

## Problem Statement
El módulo de auth guarda session tokens en localStorage. Legal flaggeó
esto como riesgo de compliance. Hay que migrar a httpOnly cookies antes
del cierre del Q2.

## Solution
Los usuarios podrán autenticarse. Las sesiones persistirán entre recargas.
Un logout invalida la sesión en servidor y cliente. Los tokens nunca tocan
el JavaScript del frontend. Requests con token expirado reciben 401 y el
cliente redirige a login.

## User Stories
Como usuario registrado, cuando envío credenciales válidas, accedo al sistema
y mi sesión persiste 24 horas.

Como sistema, cuando recibo un request con token expirado, devuelvo 401 y el
cliente redirige a login.

Como usuario autenticado, cuando hago logout, mi sesión queda invalidada en
el servidor y el token anterior no puede reutilizarse.

## Constraints
- Tokens nunca en localStorage ni sessionStorage
- Invalidación de sesión en servidor (no solo borrar cookie en cliente)
- httpOnly cookies únicamente

## Acceptance Criteria
- [ ] Login con credenciales válidas → 200 + cookie httpOnly
- [ ] Login con credenciales inválidas → 401 + mensaje genérico
- [ ] Request con token expirado → 401
- [ ] Logout → sesión invalidada, reutilizar token → 401
- [ ] Tests de integración cubren todos los escenarios anteriores

## Out of Scope
- OAuth / login social (fase 2)
- 2FA (fase 3)
- Rate limiting en login (issue separado)

## Assumptions
- El proyecto ya tiene Prisma configurado con tabla `users`
- El frontend es React SPA + API separada — no SSR
- No hay tests de integración actualmente — se crean en este PRD

## Implementation Notes
- Considera JWT vs sesiones en base de datos según si necesitas invalidación real en servidor
- Si usas JWT, evalúa si necesitas blacklist para invalidación
