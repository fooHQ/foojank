# foojank

**foojank** is a cross-platform C2 framework suitable for CTF teams. Vessel, foojank's agent, contains a runtime scripting
engine which allows the operators to execute custom payloads written in the scripting language
[Risor](https://github.com/risor-io/risor). The framework comes equipped with continuously expanding arsenal of scripts, though
teams are encouraged to write their own scripts to keep that competitive edge. The existing capabilities of the engine can be further
extended with modules. 

Unlike other C2 frameworks, foojank does not have a custom C2 server. Instead, the individual parts of the framework
communicate by passing messages via a NATS server. This design fosters collaboration as connected agents are discoverable
by other team members.

## Usage

## Configuration

## Custom modules
