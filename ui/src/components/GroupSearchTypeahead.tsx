import React from "react";
import { AsyncTypeahead } from "react-bootstrap-typeahead";
import "react-bootstrap-typeahead/css/Typeahead.css";
import ProfilePicture from "@/components/ProfilePicture";
import { TranslationFunc } from "@/components/withTranslation";
import Search, { SearchOptions } from "@/types/Search";

interface Props {
  t: TranslationFunc;
  id?: string;
  inputProps?: any;
  disabled?: boolean;
  multiple: boolean;
  selected?: any[];
  defaultSelected?: any[];
  onChange: (selected: any[]) => void;
}

interface State {
  options: any[];
  loading: boolean;
}

class GroupSearchTypeahead extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      options: [],
      loading: false,
    };
  }

  filterBy = () => {
    return true;
  };

  handleSearch = (query: string) => {
    this.setState({ loading: true });
    const options = new SearchOptions();
    options.includeGroups = true;
    options.keyword = query ? query : "";
    Search.search(options).then((res) => {
      this.setState({
        options: res.groups,
        loading: false,
      });
    });
  };

  render() {
    return (
      <AsyncTypeahead
        disabled={this.props.disabled}
        filterBy={this.filterBy}
        id={this.props.id}
        inputProps={this.props.inputProps}
        isLoading={this.state.loading}
        labelKey="name"
        multiple={this.props.multiple}
        minLength={3}
        selected={this.props.selected}
        defaultSelected={this.props.defaultSelected}
        onChange={this.props.onChange}
        onSearch={this.handleSearch}
        options={this.state.options}
        placeholder={this.props.t("searchForGroup")}
        renderMenuItemChildren={(option: any) => (
          <div className="d-flex">
            <ProfilePicture width={24} height={24} />
            <span style={{ marginLeft: "10px" }}>{option.name}</span>
          </div>
        )}
      />
    );
  }
}

export default GroupSearchTypeahead;
