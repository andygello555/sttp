git archive --add-file=documents/context_free_languages_parser_generators.pdf --add-file=documents/derivations_idioms_associativity_priority.pdf --add-file=documents/full.pdf --add-file=documents/interim.pdf --add-file=documents/project_plan.pdf --add-file=documents/specification_for_language.pdf -o tmp.zip HEAD &&
mkdir tmp &&
mv tmp.zip tmp &&
cd tmp &&
unzip tmp.zip &&
rm tmp.zip &&
cp context_free_languages_parser_generators.pdf reports/context_free_grammars_and_manual_procedures &&
cp derivations_idioms_associativity_priority.pdf reports/derivations_idioms_associativity_priority &&
cp full.pdf reports/full &&
cp interim.pdf reports/interim &&
cp project_plan.pdf reports/ &&
cp specification_for_language.pdf reports/specification_for_language &&
mkdir documents &&
mv context_free_languages_parser_generators.pdf documents &&
mv derivations_idioms_associativity_priority.pdf documents &&
mv full.pdf documents &&
mv interim.pdf documents &&
mv project_plan.pdf documents &&
mv specification_for_language.pdf documents &&
zip -r $1.zip . &&
mv $1.zip ../ &&
cd ../ &&
rm -r tmp
